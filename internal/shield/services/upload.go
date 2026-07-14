package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/tasks"
	"strings"

	"github.com/google/uuid"
)

type UploadServiceInt interface {
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
	uploader            *s3.Uploader
	logUploadRepository repositories.LogUploadRepositoryInt
	taskClient          *tasks.TaskClient
	dataSourceRepo      repositories.DataSourceRepositoryInt
}

func NewUploadService(
	uploader *s3.Uploader,
	logUploadRepository repositories.LogUploadRepositoryInt,
	taskClient *tasks.TaskClient,
	dataSourceRepo repositories.DataSourceRepositoryInt,
) UploadServiceInt {
	return &UploadService{
		uploader:            uploader,
		logUploadRepository: logUploadRepository,
		taskClient:          taskClient,
		dataSourceRepo:      dataSourceRepo,
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
	fileClass, err := MapContentTypeToFileClass(contentType, filename)
	if err != nil {
		return nil, fmt.Errorf("map content type to file class: %w", err)
	}

	if err := ValidateFileExtension(filename, fileClass); err != nil {
		return nil, fmt.Errorf("validate file extension: %w", err)
	}

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
		UserID:         parsedUUID,
		OrganizationID: orgID,
		S3Key:          key,
		FileClass:      fileClass,
	})
	if err != nil {
		return nil, err
	}

	result := &models.PresignUploadResponse{
		Key:       key,
		ExpiresAt: *expiresAt,
		Post:      *presignResult,
	}

	return result, nil
}

func (s *UploadService) ValidateUploadComplete(
	ctx context.Context,
	rc *middlewares.RequestContext,
	req *requests.UploadCompleteRequest,
) (*models.LogFile, error) {
	orgID, err := uuid.Parse(rc.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("parse organization id: %w", err)
	}
	ds, err := s.ResolveDataSource(ctx, orgID, req.Metadata)
	if err != nil {
		return nil, err
	}

	req.Metadata.SourceID = &ds.ID

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

	lf, err := s.logUploadRepository.GetByS3Key(ctx, req.Key)
	if err != nil {
		return nil, fmt.Errorf("get log file by S3 key: %w", err)
	}

	if lf.OrganizationID.String() != rc.OrganizationID {
		_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "unauthorized user")
		return nil, fmt.Errorf("unauthorized")
	}

	if lf.Status != "pending" {
		return nil, fmt.Errorf("invalid log file status")
	}

	lf, err = s.logUploadRepository.Confirm(ctx, models.ConfirmLogFileParams{
		S3Key:        req.Key,
		SourceType:   req.Metadata.SourceType,
		SourceID:     req.Metadata.SourceID,
		Description:  &req.Metadata.Description,
		Category:     &req.Metadata.Category,
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

	// err = s.taskClient.EnqueueSubmitLogFileForAnalysis(ctx, models.SubmitLogFileInput{
	// 	LogFileID:      lf.ID,
	// 	S3Key:          lf.S3Key,
	// 	FileClass:      lf.FileClass,
	// 	SourceType:     sourceType,
	// 	OrganizationID: lf.OrganizationID,
	// 	UserID:         parsedUserID,
	// })
	// if err != nil {
	// 	_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "failed to enqueue analysis task")
	// 	return nil, fmt.Errorf("enqueue analysis task: %w", err)
	// }

	err = s.taskClient.EnqueueSubmitLogFileForProcessing(ctx, models.SubmitLogFileInput{
		LogFileID:      lf.ID,
		S3Key:          lf.S3Key,
		FileClass:      lf.FileClass,
		SourceType:     sourceType,
		OrganizationID: lf.OrganizationID,
		UserID:         parsedUserID,
		SourceID:       lf.SourceID,
	})
	if err != nil {
		_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "failed to enqueue analysis task")
		return nil, fmt.Errorf("enqueue analysis task: %w", err)
	}

	return lf, nil
}

func MapContentTypeToFileClass(contentType string, filename string) (models.FileClass, error) {
	switch contentType {
	case "text/csv":
		return models.FileClassCSV, nil
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return models.FileClassXLSX, nil
	case "application/json":
		return models.FileClassJSON, nil
	case "application/vnd.tcpdump.pcap", "application/octet-stream":
		if strings.HasSuffix(filename, ".pcap") {
			return models.FileClassPCAP, nil
		}
	case "text/plain":
		if strings.HasSuffix(filename, ".log") {
			return models.FileClassLog, nil
		}
	}

	return "", fmt.Errorf("unsupported content type: %s", contentType)
}

func ValidateFileExtension(filename string, class models.FileClass) error {
	switch class {
	case models.FileClassCSV:
		if !strings.HasSuffix(filename, ".csv") {
			return errors.New("invalid file extension for csv")
		}
	case models.FileClassXLSX:
		if !strings.HasSuffix(filename, ".xlsx") {
			return errors.New("invalid file extension for xlsx")
		}
	case models.FileClassJSON:
		if !strings.HasSuffix(filename, ".json") {
			return errors.New("invalid file extension for json")
		}
	case models.FileClassLog:
		if !strings.HasSuffix(filename, ".log") {
			return errors.New("invalid file extension for log")
		}
	case models.FileClassPCAP:
		if !strings.HasSuffix(filename, ".pcap") {
			return errors.New("invalid file extension for pcap")
		}
	}
	return nil
}

func (s *UploadService) ResolveDataSource(
	ctx context.Context,
	orgID uuid.UUID,
	meta requests.LogUploadMetadata,
) (*models.DataSource, error) {

	// 1. If user provided SourceID → fetch it
	if meta.SourceID != nil {
		ds, err := s.dataSourceRepo.GetDataSourceByID(ctx, *meta.SourceID, orgID)
		if err != nil {
			return nil, err
		}
		return ds, nil
	}

	// 2. Otherwise → create one automatically

	name := meta.Category
	provider := meta.SourceType // or meta.Category

	ds := &models.DataSource{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Type:           "log",
		Provider:       &provider,
		Status:         "active",
		Metadata:       buildMetadata(meta),
	}

	err := s.dataSourceRepo.CreateDataSource(ctx, ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func buildMetadata(meta requests.LogUploadMetadata) []byte {
	m := map[string]interface{}{
		"source_type": meta.SourceType,
		"category":    meta.Category,
		"context":     meta.AppOrContext,
		"host":        meta.Host,
		"index_name":  meta.IndexName,
	}

	b, _ := json.Marshal(m)
	return b
}
