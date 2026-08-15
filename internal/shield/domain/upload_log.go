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
	ID               uuid.UUID
	UserID           uuid.UUID
	OrganizationID   uuid.UUID
	S3Key            string
	FileClass        FileClass
	EventCount       *int
	ProcessedAt      *time.Time
	SourceType       *string
	SourceID         *uuid.UUID
	Description      *string
	Category         *string
	AppOrContext     *string
	Status           string
	ErrorMessage     *string
	DetectedType     *DetectedType
	UserSelectedType *DetectedType
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
