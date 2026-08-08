package ai_detector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
	"sage-backend/internal/shield/ports/dto"
	"time"

	"github.com/google/uuid"
)

type ThreatDetectorInt interface {
	SubmitLogFileForAnalysis(
		ctx context.Context,
		input domain.SubmitLogFileForAnalysis,
	) (*dto.SubmitLogFileResult, error)
}

type ThreatDetector struct {
	client        *AIDetectorClient
	s3Uploader    *s3.Uploader
	logUploadRepo outbound.LogUploadRepository
	analysisRepo  outbound.AnalysisRepository
}

func NewThreatDetector(
	client *AIDetectorClient,
	s3Uploader *s3.Uploader,
	logUploadRepo outbound.LogUploadRepository,
	analysisRepo outbound.AnalysisRepository,
) ThreatDetectorInt {
	return &ThreatDetector{
		client:        client,
		s3Uploader:    s3Uploader,
		logUploadRepo: logUploadRepo,
		analysisRepo:  analysisRepo,
	}
}

func (td *ThreatDetector) SubmitLogFileForAnalysis(
	ctx context.Context,
	input domain.SubmitLogFileForAnalysis,
) (*dto.SubmitLogFileResult, error) {
	if input.LogFileID == uuid.Nil {
		return nil, fmt.Errorf("log file id is required")
	}

	if input.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization id is required")
	}

	if existing, err := td.analysisRepo.GetByLogFileID(ctx, input.LogFileID); err == nil && existing != nil {
		return &dto.SubmitLogFileResult{
			JobID:       existing.ID.String(),
			SubmittedAt: existing.CreatedAt,
		}, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing analysis: %w", err)
	}

	health, err := td.client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("detector health check failed: %w", err)
	}

	if health.Status != "ok" {
		return nil, fmt.Errorf("detector is not healthy: %s", health.Service)
	}

	analysis, err := td.client.DetectFileThreats(ctx, input.FileReader, input.FileName)
	if err != nil {
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "threat detection failed")
		return nil, fmt.Errorf("detect threats: %w", err)
	}

	payload := domain.CreateAnalysisParams{
		LogFileID:      &input.LogFileID,
		RequestType:    domain.AnalysisRequestTypeFile,
		LogType:        input.FileClass,
		Approach:       analysis.Approach,
		Overall:        analysis.Overall,
		Summary:        analysis.Summary,
		Outcome:        analysis.Outcome,
		Threats:        analysis.Threats,
		OrganizationID: input.OrganizationID,
	}

	result, err := td.analysisRepo.RecordAnalysis(ctx, &payload)
	if err != nil {
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "failed to persist analysis result")
		return nil, fmt.Errorf("record analysis: %w", err)
	}

	if err := td.logUploadRepo.MarkSubmitted(ctx, input.S3Key); err != nil {
		_ = td.logUploadRepo.MarkFailed(ctx, input.S3Key, "failed to mark file as submitted")
		return nil, fmt.Errorf("mark submitted: %w", err)
	}

	return &dto.SubmitLogFileResult{
		JobID:       result.ID.String(),
		SubmittedAt: time.Now().UTC(),
	}, nil
}
