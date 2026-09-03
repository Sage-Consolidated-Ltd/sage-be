package dto

import (
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
)

// IngestLogRequest represents a single log ingestion
type IngestLogRequest struct {
	SourceID      string                 `json:"source_id" validate:"required,uuid"`
	SourceEventID string                 `json:"source_event_id,omitempty"`
	EventType     string                 `json:"event_type" validate:"required"`
	EventCategory string                 `json:"event_category" validate:"required"`
	Severity      types.Severity         `json:"severity" validate:"required,oneof=low medium high critical"`
	ActorEmail    string                 `json:"actor_email,omitempty"`
	ActorUsername string                 `json:"actor_username,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	GeoCountry    string                 `json:"geo_country,omitempty"`
	GeoCity       string                 `json:"geo_city,omitempty"`
	OccurredAt    time.Time              `json:"occurred_at" validate:"required"`
	RawPayload    map[string]interface{} `json:"raw_payload" validate:"required"`
}

// BulkIngestLogsRequest represents bulk log ingestion
type BulkIngestLogsRequest struct {
	SourceID string              `json:"source_id" validate:"required,uuid"`
	Events   []*IngestLogRequest `json:"events" validate:"required,min=1,dive"`
}

type IntegrateDataSource struct {
	Provider       string                 `json:"provider" validate:"required"`
	ConnectionType string                 `json:"connection_type" validate:"required"`
	Config         map[string]interface{} `json:"config" validate:"required"`
	Credentials    map[string]interface{} `json:"credentials" validate:"required"`
}

type SecurityEventResponse struct {
	ID                string          `json:"id"`
	SourceID          string          `json:"source_id"`
	Source            string          `json:"source"`
	SourceEventID     *string         `json:"source_event_id,omitempty"`
	EventType         string          `json:"event_type"`
	EventCategory     string          `json:"event_category"`
	Severity          string          `json:"severity"`
	ActorEmail        *string         `json:"actor_email,omitempty"`
	ActorUsername     *string         `json:"actor_username,omitempty"`
	IPAddress         *string         `json:"ip_address,omitempty"`
	GeoCountry        *string         `json:"geo_country,omitempty"`
	GeoCity           *string         `json:"geo_city,omitempty"`
	RawPayload        db.JSONMap      `json:"raw_payload"`
	NormalizedPayload db.JSONMap      `json:"normalized_payload"`
	ParserID          *string         `json:"parser_id,omitempty"`
	ParseStatus       string          `json:"parse_status"`
	ParseErrors       db.JSONMapSlice `json:"parse_errors,omitempty"`
	OccurredAt        time.Time       `json:"occurred_at"`
	IngestedAt        time.Time       `json:"ingested_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	IngestionType     string          `json:"ingestion_type,omitempty"`
	FileID            *string         `json:"file_id,omitempty"`
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

