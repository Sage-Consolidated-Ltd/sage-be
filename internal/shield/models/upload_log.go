package models

import (
	"sage-backend/internal/shared/storage/s3"
	"time"

	"github.com/google/uuid"
)

type PresignUploadResponse struct {
	Key       string        `json:"key"`
	ExpiresAt time.Time     `json:"expires_at"`
	Post      s3.PresignedPost `json:"post"`
}

type LogFile struct {
	ID             uuid.UUID  `db:"id"`
	UserID         uuid.UUID  `db:"user_id"`
	OrganizationID uuid.UUID  `db:"organization_id"`
	S3Key          string     `db:"s3_key"`
	FileClass      string     `db:"file_class"`
	SourceType     *string     `db:"source_type"`
	SourceID       *uuid.UUID `db:"source_id"`
	Description    *string    `db:"description"`
	Category       *string    `db:"category"`
	AppOrContext   *string    `db:"app_or_context"`
	Status         string     `db:"status"`
	ErrorMessage   *string    `db:"error_message"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

type CreateLogFileParams struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	S3Key          string
	FileClass      string
}

type ConfirmLogFileParams struct {
	S3Key       string
	SourceType  string
	SourceID    *uuid.UUID
	Description *string
	Category *string
	AppOrContext *string
}