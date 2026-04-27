package requests

import (
	"time"

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
