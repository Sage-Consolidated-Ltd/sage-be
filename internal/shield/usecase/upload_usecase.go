package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type UploadService struct {
	uploader            *s3.Uploader
	logUploadRepository outbound.LogUploadRepository
	taskPublisher       outbound.TaskPublisherInt
	dataSourceRepo      outbound.DataSourceRepository
}

func NewUploadService(
	uploader *s3.Uploader,
	logUploadRepository outbound.LogUploadRepository,
	taskPublisher outbound.TaskPublisherInt,
	dataSourceRepo outbound.DataSourceRepository,
) inbound.UploadUseCase {
	return &UploadService{
		uploader:            uploader,
		logUploadRepository: logUploadRepository,
		taskPublisher:       taskPublisher,
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
) (*dto.PresignUploadResponse, error) {
	if s.uploader == nil {
		return nil, errors.New("s3 uploader is not configured")
	}

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

	parsedUUID, err := uuid.Parse(rc.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user UUID: %w", err)
	}

	// save pending log file record
	_, err = s.logUploadRepository.CreatePending(ctx, domain.CreateLogFileParams{
		UserID:         parsedUUID,
		OrganizationID: orgID,
		S3Key:          key,
		FileClass:      fileClass,
	})
	if err != nil {
		return nil, err
	}

	result := &dto.PresignUploadResponse{
		Key:       key,
		ExpiresAt: *expiresAt,
		Post:      *presignResult,
	}

	return result, nil
}

func (s *UploadService) ValidateUploadComplete(
	ctx context.Context,
	rc *middlewares.RequestContext,
	req *dto.UploadCompleteRequest,
) (*domain.LogFile, error) {
	if s.uploader == nil {
		return nil, errors.New("s3 uploader is not configured")
	}

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

	lf, err = s.logUploadRepository.Confirm(ctx, domain.ConfirmLogFileParams{
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

	if s.taskPublisher != nil {
		err = s.taskPublisher.EnqueueSubmitLogFile(ctx, dto.SubmitLogFileInput{
			LogFileID:      lf.ID,
			S3Key:          lf.S3Key,
			FileClass:      lf.FileClass,
			SourceType:     sourceType,
			OrganizationID: lf.OrganizationID,
			UserID:         parsedUserID,
			SourceID:       lf.SourceID,
		})
		if err != nil {
			_ = s.logUploadRepository.MarkFailed(ctx, req.Key, "failed to enqueue processing task")
			return nil, fmt.Errorf("enqueue processing task: %w", err)
		}
	}

	return lf, nil
}

func MapContentTypeToFileClass(contentType string, filename string) (domain.FileClass, error) {
	switch contentType {
	case "text/csv":
		return domain.FileClassCSV, nil
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return domain.FileClassXLSX, nil
	case "application/json":
		return domain.FileClassJSON, nil
	case "application/vnd.tcpdump.pcap", "application/octet-stream":
		if strings.HasSuffix(filename, ".pcap") {
			return domain.FileClassPCAP, nil
		}
		if strings.HasSuffix(filename, ".evt") || strings.HasSuffix(filename, ".evtx") {
			return domain.FileClassLog, nil
		}
	case "text/plain":
		if strings.HasSuffix(filename, ".log") || strings.HasSuffix(filename, ".txt") {
			return domain.FileClassLog, nil
		}
	}

	// Fallback to extension check if content-type is generic
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return domain.FileClassCSV, nil
	case ".xlsx":
		return domain.FileClassXLSX, nil
	case ".json":
		return domain.FileClassJSON, nil
	case ".log", ".txt", ".evt", ".evtx":
		return domain.FileClassLog, nil
	case ".pcap":
		return domain.FileClassPCAP, nil
	}

	return "", fmt.Errorf("unsupported content type: %s", contentType)
}

func ValidateFileExtension(filename string, class domain.FileClass) error {
	switch class {
	case domain.FileClassCSV:
		if !strings.HasSuffix(filename, ".csv") {
			return errors.New("invalid file extension for csv")
		}
	case domain.FileClassXLSX:
		if !strings.HasSuffix(filename, ".xlsx") {
			return errors.New("invalid file extension for xlsx")
		}
	case domain.FileClassJSON:
		if !strings.HasSuffix(filename, ".json") {
			return errors.New("invalid file extension for json")
		}
	case domain.FileClassLog:
		if !strings.HasSuffix(filename, ".log") && !strings.HasSuffix(filename, ".txt") && !strings.HasSuffix(filename, ".evt") && !strings.HasSuffix(filename, ".evtx") {
			return errors.New("invalid file extension for log")
		}
	case domain.FileClassPCAP:
		if !strings.HasSuffix(filename, ".pcap") {
			return errors.New("invalid file extension for pcap")
		}
	}
	return nil
}

func (s *UploadService) ResolveDataSource(
	ctx context.Context,
	orgID uuid.UUID,
	meta dto.LogUploadMetadata,
) (*domain.DataSource, error) {
	if meta.SourceID != nil {
		ds, err := s.dataSourceRepo.GetDataSourceByID(ctx, *meta.SourceID, orgID)
		if err != nil {
			return nil, err
		}
		return ds, nil
	}

	name := meta.Category
	if name == "" {
		name = meta.SourceType
	}
	provider := meta.SourceType

	metaBytes, _ := json.Marshal(map[string]interface{}{
		"source_type": meta.SourceType,
		"category":    meta.Category,
		"context":     meta.AppOrContext,
		"host":        meta.Host,
		"index_name":  meta.IndexName,
	})

	ds := &domain.DataSource{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Type:           "log",
		Provider:       &provider,
		Status:         "active",
		Metadata:       metaBytes,
	}

	err := s.dataSourceRepo.CreateDataSource(ctx, ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
