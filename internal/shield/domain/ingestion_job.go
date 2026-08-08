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
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	SourceID        *uuid.UUID
	Status          JobStatus
	JobType         JobType
	EventsProcessed int64
	EventsFailed    int64
	ErrorMessage    *string
	Metadata        map[string]interface{}
	StartedAt       *time.Time
	CompletedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
