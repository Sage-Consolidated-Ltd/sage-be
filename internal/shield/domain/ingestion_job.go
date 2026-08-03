package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	JobTypeIngest      JobType = "ingest"
	JobTypeSync        JobType = "sync"
	JobTypeQualityScan JobType = "quality_scan"
	JobTypeValidation  JobType = "validation"
)

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

func (j JobStatus) IsValid() bool {
	switch j {
	case JobStatusQueued, JobStatusRunning, JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	default:
		return false
	}
}

type IngestionJob struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	OrganizationID  uuid.UUID              `db:"organization_id" json:"organization_id"`
	SourceID        *uuid.UUID             `db:"source_id,omitempty" json:"source_id,omitempty"`
	Status          JobStatus              `db:"status" json:"status"`
	JobType         JobType                `db:"job_type" json:"job_type"`
	EventsProcessed int64                  `db:"events_processed" json:"events_processed"`
	EventsFailed    int64                  `db:"events_failed" json:"events_failed"`
	ErrorMessage    *string                `db:"error_message,omitempty" json:"error_message,omitempty"`
	Metadata        map[string]interface{} `db:"metadata" json:"metadata"`
	StartedAt       *time.Time             `db:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt     *time.Time             `db:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

type IngestionJobResponse struct {
	ID              string                 `json:"id"`
	OrganizationID  string                 `json:"organization_id"`
	SourceID        *string                `json:"source_id,omitempty"`
	Status          string                 `json:"status"`
	JobType         string                 `json:"job_type"`
	EventsProcessed int64                  `json:"events_processed"`
	EventsFailed    int64                  `json:"events_failed"`
	ErrorMessage    *string                `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

func (i *IngestionJob) ToResponse() *IngestionJobResponse {
	sid := ""
	if i.SourceID != nil {
		sid = i.SourceID.String()
	}
	return &IngestionJobResponse{
		ID:              i.ID.String(),
		OrganizationID:  i.OrganizationID.String(),
		SourceID:        &sid,
		Status:          string(i.Status),
		JobType:         string(i.JobType),
		EventsProcessed: i.EventsProcessed,
		EventsFailed:    i.EventsFailed,
		ErrorMessage:    i.ErrorMessage,
		Metadata:        i.Metadata,
		StartedAt:       i.StartedAt,
		CompletedAt:     i.CompletedAt,
		CreatedAt:       i.CreatedAt,
	}
}
