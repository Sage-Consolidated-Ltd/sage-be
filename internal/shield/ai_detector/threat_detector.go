package ai_detector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"
	"time"

	"github.com/google/uuid"
)

type ThreatDetectorInt interface {
	SubmitLogFileForAnalysis(
		ctx context.Context,
		input models.SubmitLogFileForAnalysis,
	) (*models.SubmitLogFileResult, error)
}

type ThreatDetector struct {
	client     *AIDetectorClient
	s3Uploader *s3.Uploader
	logUploadRepo repositories.LogUploadRepositoryInt
	analysisRepo repositories.AnalysisRepositoryInt
}

func NewThreatDetector(
	client *AIDetectorClient, 
	s3Uploader *s3.Uploader, 
	logUploadRepo repositories.LogUploadRepositoryInt,
	analysisRepo repositories.AnalysisRepositoryInt,
	) ThreatDetectorInt {
	return &ThreatDetector{
		client:     client,
		s3Uploader: s3Uploader,
		logUploadRepo: logUploadRepo,
		analysisRepo: analysisRepo,
	}
}


func (td *ThreatDetector) SubmitLogFileForAnalysis(
	ctx context.Context,
	input models.SubmitLogFileForAnalysis,
) (*models.SubmitLogFileResult, error) {
	fmt.Printf("Submitting log file for analysis: %s\n", input.LogFileID)
	fmt.Printf("Input details: OrganizationID: %s, Filename: %s, FileClass: %s, LogFileID: %s\n", input.OrganizationID, input.FileName, input.FileClass, input.LogFileID)
	if input.LogFileID == uuid.Nil {
		return nil, fmt.Errorf("log file id is required")
	}

	if input.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization id is required")
	}

	fmt.Printf("Checking for existing analysis for log file: %s\n", input.LogFileID)
	if existing, err := td.analysisRepo.GetByLogFileID(ctx, input.LogFileID); err == nil && existing != nil {
		return &models.SubmitLogFileResult{
			JobID:       existing.ID.String(),
			SubmittedAt: existing.CreatedAt,
		}, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("Error checking existing analysis for log file %s: %v\n", input.LogFileID, err)
		return nil, fmt.Errorf("check existing analysis: %w", err)
	}

	// confirm health of the AI detector service before proceeding
	fmt.Printf("Checking health of AI detector service...\n")
	health, err := td.client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("detector health check failed: %w", err)
	}

	if health.Status != "ok" {
		return nil, fmt.Errorf("detector is not healthy: %s", health.Service)
	}

	fmt.Printf("AI detector service is healthy\n")
	
	fmt.Printf("Detecting threats in log file: %s\n", input.FileName)
	analysis, err := td.client.DetectFileThreats(ctx, input.FileReader, input.FileName)
	if err != nil {
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "threat detection failed")
		return nil, fmt.Errorf("detect threats: %w", err)
	}

	payload := models.CreateAnalysisParams{
		LogFileID: &input.LogFileID,
		RequestType: models.AnalysisRequestTypeFile,
		LogType: input.FileClass,
		Approach: analysis.Approach,
		Overall: analysis.Overall,
		Summary: analysis.Summary,
		Outcome: analysis.Outcome,
		Threats: analysis.Threats,
		OrganizationID: input.OrganizationID,
	}

	fmt.Printf("Recording analysis for log file: %s\n", input.LogFileID)
	result, err := td.analysisRepo.RecordAnalysis(ctx, &payload)
	if err != nil {
		fmt.Printf("An error occured persisting analysis: %v\n", err)
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "failed to persist analysis result")
		return nil, fmt.Errorf("record analysis: %w", err)
	}

	if err := td.logUploadRepo.MarkSubmitted(ctx, input.S3Key); err != nil {
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "failed to mark file as submitted")
		return nil, fmt.Errorf("mark submitted: %w", err)
	}

	return &models.SubmitLogFileResult{
		JobID:       result.ID.String(),
		SubmittedAt: time.Now().UTC(),
	}, nil
}