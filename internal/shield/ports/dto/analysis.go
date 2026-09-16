package dto

import (
	"time"

	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

type SubmitLogFileInput struct {
	LogFileID      uuid.UUID        `json:"log_file_id"`
	S3Key          string           `json:"s3_key"`
	FileClass      domain.FileClass `json:"file_class"`
	SourceType     string           `json:"source_type"`
	SourceID       *uuid.UUID       `json:"source_id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	UserID         uuid.UUID        `json:"user_id"`
}

type SubmitLogFileResult struct {
	JobID       string    `json:"job_id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type CheckHealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}
