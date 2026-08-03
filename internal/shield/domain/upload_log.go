package domain

import (
	"sage-backend/internal/shared/storage/s3"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusUploaded   Status = "uploaded"
	StatusProcessing Status = "processing"
	StatusParsed     Status = "parsed"
	StatusAnalyzed   Status = "analyzed"
	StatusFailed     Status = "failed"
	StatusRetrying   Status = "retrying"
)

type FileClass string

const (
	FileClassCSV  FileClass = "csv"
	FileClassXLSX FileClass = "xlsx"
	FileClassLog  FileClass = "log"
	FileClassJSON FileClass = "json"
	FileClassPCAP FileClass = "pcap"
)

type DetectedType string

const (
	DetectedWindowsEventLog DetectedType = "windows_event_log"
	DetectedLinuxSyslog     DetectedType = "linux_syslog"
	DetectedCSVStructured   DetectedType = "csv_structured"
	DetectedXLSXStructured  DetectedType = "xlsx_structured"
	DetectedUnknown         DetectedType = "unknown"
)

type PresignUploadResponse struct {
	Key       string           `json:"key"`
	ExpiresAt time.Time        `json:"expires_at"`
	Post      s3.PresignedPost `json:"post"`
}

type LogFile struct {
	ID               uuid.UUID     `db:"id"`
	UserID           uuid.UUID     `db:"user_id"`
	OrganizationID   uuid.UUID     `db:"organization_id"`
	S3Key            string        `db:"s3_key"`
	FileClass        FileClass     `db:"file_class"`
	EventCount       *int          `db:"event_count"`
	ProcessedAt      *time.Time    `db:"processed_at"`
	SourceType       *string       `db:"source_type"`
	SourceID         *uuid.UUID    `db:"source_id"`
	Description      *string       `db:"description"`
	Category         *string       `db:"category"`
	AppOrContext     *string       `db:"app_or_context"`
	Status           string        `db:"status"`
	ErrorMessage     *string       `db:"error_message"`
	DetectedType     *DetectedType `db:"detected_type"`
	UserSelectedType *DetectedType `db:"user_selected_type"`
	CreatedAt        time.Time     `db:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at"`
}

func (f *LogFile) EffectiveType() DetectedType {
	if f.UserSelectedType != nil && *f.UserSelectedType != "" {
		return *f.UserSelectedType
	}
	if f.DetectedType != nil {
		return *f.DetectedType
	}
	return DetectedUnknown
}

type CreateLogFileParams struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	S3Key          string
	FileClass      FileClass
}

type ConfirmLogFileParams struct {
	S3Key        string
	SourceType   string
	SourceID     *uuid.UUID
	Description  *string
	Category     *string
	AppOrContext *string
}
