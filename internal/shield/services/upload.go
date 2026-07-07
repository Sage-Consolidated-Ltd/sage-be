package services

import (
	"context"
	"fmt"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/tasks"

	"github.com/google/uuid"
)

type UploadServiceInt interface{
	GetUploadURL(
	ctx context.Context,
	rc middlewares.RequestContext,
	orgID uuid.UUID,
	filename string,
	contentType string,
	sizeBytes int64,
	) (*models.PresignUploadResponse, error)
	
	ValidateUploadComplete(
	ctx context.Context,
	rc *middlewares.RequestContext,
	req *requests.UploadCompleteRequest,
	) (*models.LogFile, error)
}

type UploadService struct {
	uploader *s3.Uploader
	logUploadRepository repositories.LogUploadRepositoryInt
	taskClient *tasks.TaskClient
}

func NewUploadService(
	uploader *s3.Uploader,
	logUploadRepository repositories.LogUploadRepositoryInt,
	taskClient *tasks.TaskClient,
) UploadServiceInt {
	return &UploadService{
		uploader: uploader,
		logUploadRepository: logUploadRepository,
		taskClient: taskClient,
	}
}

func (s *UploadService) GetUploadURL(
	ctx context.Context,
	rc middlewares.RequestContext,
	orgID uuid.UUID,
	filename string,
	contentType string,
	sizeBytes int64,
) (*models.PresignUploadResponse, error) {
	presignResult, key, expiresAt, err := s.uploader.PresignUploadPost(
		ctx,
		rc,
		filename,
		sizeBytes,
	)

	if err != nil {
		return nil, err
	}

	userID := rc.UserID
	fmt.Println(userID)
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UUID: %v", err)
	}

	// save pending 
	_, err = s.logUploadRepository.CreatePending(ctx, models.CreateLogFileParams{
		UserID: parsedUUID,
		OrganizationID: orgID,
		S3Key: key,
		FileClass: contentType,
	})
	if err != nil {
		return nil, err
	}

	result := &models.PresignUploadResponse{
		Key: 	 key,
		ExpiresAt: *expiresAt,
		Post: *presignResult,
	}

	return result, nil
}

func (s *UploadService) ValidateUploadComplete(
	ctx context.Context,
	rc *middlewares.RequestContext,
	req *requests.UploadCompleteRequest,
) (*models.LogFile, error) {
	head, err := s.uploader.HeadObject(ctx, req.Key)
	if err != nil {
		if s3.NotFoundError(err) {
			return nil, fmt.Errorf("file not found in S3, upload may have failed")
		}
		return nil, err
	}

	if head.ETag == nil {
		return nil, fmt.Errorf("uploaded file has no ETag")
	}

	if match := s.uploader.ValidateETag(*head.ETag, req.ETag); !match {
		return nil, fmt.Errorf("ETag mismatch: file may be corrupted or incomplete")
	}

	lf, err := s.logUploadRepository.Confirm(ctx, models.ConfirmLogFileParams{
		S3Key:       req.Key,
		SourceType:  req.Metadata.SourceType,
		SourceID:    req.Metadata.SourceID,
		Description: &req.Metadata.Description,
		Category: &req.Metadata.Category,
		AppOrContext: &req.Metadata.AppOrContext,
	})
	if err != nil {
		return nil, fmt.Errorf("confirm upload: %w", err)
	}

	parsedUserID, err := uuid.Parse(rc.UserID)
	if err != nil {
		_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "invalid user id in upload context")
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	sourceType := ""
	if lf.SourceType != nil {
		sourceType = *lf.SourceType
	}

	err = s.taskClient.EnqueueSubmitLogFileForAnalysis(ctx, models.SubmitLogFileInput{
		LogFileID:      lf.ID,
		S3Key:          lf.S3Key,
		FileClass:      lf.FileClass,
		SourceType:     sourceType,
		OrganizationID: lf.OrganizationID,
		UserID:         parsedUserID,
	})
	if err != nil {
		_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "failed to enqueue analysis task")
		return nil, fmt.Errorf("enqueue analysis task: %w", err)
	}

	return lf, nil
}