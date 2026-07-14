package requests

import "github.com/google/uuid"

// UploadLogRequest contains the file metadata required to initiate a presigned upload.
type UploadLogRequest struct {
	// Original filename including extension e.g. windows_security.csv
	Filename string `json:"filename" validate:"required"`
	// MIME type of the file e.g. text/csv
	ContentType string `json:"contentType" validate:"required"`
	// File size in bytes — must match the actual file size uploaded to S3
	Size int64 `json:"size" validate:"required"`
}

// UploadCompleteRequest confirms a completed S3 upload.
type UploadCompleteRequest struct {
	// S3 object key returned by the presign endpoint
	Key string `json:"key"      validate:"required" example:"uploads/pending/csv/428d4fcc-.../abc-123.csv"`
	// ETag header value from the S3 POST response — include surrounding quotes
	ETag     string            `json:"etag"     validate:"required" example:"\"7f5d37276b649790cfd8f11ee5946e02\""`
	Metadata LogUploadMetadata `json:"metadata" validate:"required"`
}

// LogUploadMetadata contains user-provided context about the uploaded log file.
type LogUploadMetadata struct {
	// Log source type e.g. firewall, nginx, windows_security
	SourceType string `json:"source_type"    validate:"required" example:"windows_security"`
	// Optional ID of an existing data source to link this log file to
	SourceID     *uuid.UUID `json:"source_id"                         example:"a1b2c3d4-..."`
	Description  string     `json:"description"                       example:"Windows security event logs"`
	Category     string     `json:"category"                          example:"windows"`
	AppOrContext string     `json:"app_or_context"                    example:"windows"`
	Host         string     `json:"host"                              example:"DESKTOP-1234567"`
	IndexName    string     `json:"index_name"                        example:"windows_security"`
}
