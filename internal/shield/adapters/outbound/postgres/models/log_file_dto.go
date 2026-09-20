package models

import (
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type LogFileDTO struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	OrganizationID   uuid.UUID  `db:"organization_id"`
	S3Key            string     `db:"s3_key"`
	FileClass        string     `db:"file_class"`
	EventCount       *int       `db:"event_count"`
	ProcessedAt      *time.Time `db:"processed_at"`
	SourceType       *string    `db:"source_type"`
	SourceID         *uuid.UUID `db:"source_id"`
	Description      *string    `db:"description"`
	Category         *string    `db:"category"`
	AppOrContext     *string    `db:"app_or_context"`
	Status           string     `db:"status"`
	ErrorMessage     *string    `db:"error_message"`
	DetectedType     *string    `db:"detected_type"`
	UserSelectedType *string    `db:"user_selected_type"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (dto *LogFileDTO) ToDomain() *domain.LogFile {
	if dto == nil {
		return nil
	}

	var detectedType *domain.DetectedType
	if dto.DetectedType != nil {
		dt := domain.DetectedType(*dto.DetectedType)
		detectedType = &dt
	}

	var userSelectedType *domain.DetectedType
	if dto.UserSelectedType != nil {
		ust := domain.DetectedType(*dto.UserSelectedType)
		userSelectedType = &ust
	}

	return &domain.LogFile{
		ID:               dto.ID,
		UserID:           dto.UserID,
		OrganizationID:   dto.OrganizationID,
		S3Key:            dto.S3Key,
		FileClass:        domain.FileClass(dto.FileClass),
		EventCount:       dto.EventCount,
		ProcessedAt:      dto.ProcessedAt,
		SourceType:       dto.SourceType,
		SourceID:         dto.SourceID,
		Description:      dto.Description,
		Category:         dto.Category,
		AppOrContext:     dto.AppOrContext,
		Status:           dto.Status,
		ErrorMessage:     dto.ErrorMessage,
		DetectedType:     detectedType,
		UserSelectedType: userSelectedType,
		CreatedAt:        dto.CreatedAt,
		UpdatedAt:        dto.UpdatedAt,
	}
}
