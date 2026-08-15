package domain

import (
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

type LogFile struct {
	ID               uuid.UUID     `db:"id"                 json:"id"`
	UserID           uuid.UUID     `db:"user_id"            json:"user_id"`
	OrganizationID   uuid.UUID     `db:"organization_id"    json:"organization_id"`
	S3Key            string        `db:"s3_key"             json:"s3_key"`
	FileClass        FileClass     `db:"file_class"         json:"file_class"`
	EventCount       *int          `db:"event_count"        json:"event_count"`
	ProcessedAt      *time.Time    `db:"processed_at"       json:"processed_at"`
	SourceType       *string       `db:"source_type"        json:"source_type"`
	SourceID         *uuid.UUID    `db:"source_id"          json:"source_id"`
	Description      *string       `db:"description"        json:"description"`
	Category         *string       `db:"category"           json:"category"`
	AppOrContext     *string       `db:"app_or_context"     json:"app_or_context"`
	Status           string        `db:"status"             json:"status"`
	ErrorMessage     *string       `db:"error_message"      json:"error_message"`
	DetectedType     *DetectedType `db:"detected_type"      json:"detected_type"`
	UserSelectedType *DetectedType `db:"user_selected_type" json:"user_selected_type"`
	CreatedAt        time.Time     `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at"         json:"updated_at"`
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
