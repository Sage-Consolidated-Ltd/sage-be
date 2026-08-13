package models

import (
	"encoding/json"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type IngestionJobDTO struct {
	ID              uuid.UUID        `db:"id"`
	OrganizationID  uuid.UUID        `db:"organization_id"`
	SourceID        *uuid.UUID       `db:"source_id"`
	Status          domain.JobStatus `db:"status"`
	JobType         domain.JobType   `db:"job_type"`
	EventsProcessed int64            `db:"events_processed"`
	EventsFailed    int64            `db:"events_failed"`
	ErrorMessage    *string          `db:"error_message"`
	Metadata        json.RawMessage  `db:"metadata"`
	StartedAt       *time.Time       `db:"started_at"`
	CompletedAt     *time.Time       `db:"completed_at"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
}

func (dto *IngestionJobDTO) ToDomain() *domain.IngestionJob {
	if dto == nil {
		return nil
	}

	var meta map[string]interface{}
	if len(dto.Metadata) > 0 {
		_ = json.Unmarshal(dto.Metadata, &meta)
	}

	return &domain.IngestionJob{
		ID:              dto.ID,
		OrganizationID:  dto.OrganizationID,
		SourceID:        dto.SourceID,
		Status:          dto.Status,
		JobType:         dto.JobType,
		EventsProcessed: dto.EventsProcessed,
		EventsFailed:    dto.EventsFailed,
		ErrorMessage:    dto.ErrorMessage,
		Metadata:        meta,
		StartedAt:       dto.StartedAt,
		CompletedAt:     dto.CompletedAt,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
	}
}
