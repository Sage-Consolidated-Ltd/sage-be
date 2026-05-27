package models

import (
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

type SecurityEvent struct {
	ID                uuid.UUID           `db:"id" json:"id"`
	OrganizationID    uuid.UUID           `db:"organization_id" json:"organization_id"`
	SourceID          uuid.UUID           `db:"source_id" json:"source_id"`
	ParserID          *uuid.UUID          `db:"parser_id,omitempty" json:"parser_id,omitempty"`
	SourceEventID     *string             `db:"source_event_id,omitempty" json:"source_event_id,omitempty"`
	Source            string              `db:"source" json:"source"`
	EventType         string              `db:"event_type" json:"event_type"`
	EventCategory     types.EventCategory `db:"event_category" json:"event_category"`
	Severity          *types.Severity     `db:"severity" json:"severity"`
	ActorEmail        *string             `db:"actor_email,omitempty" json:"actor_email,omitempty"`
	ActorUsername     *string             `db:"actor_username,omitempty" json:"actor_username,omitempty"`
	IPAddress         *string             `db:"ip_address,omitempty" json:"ip_address,omitempty"`
	GeoCountry        *string             `db:"geo_country,omitempty" json:"geo_country,omitempty"`
	GeoCity           *string             `db:"geo_city,omitempty" json:"geo_city,omitempty"`
	RawPayload        db.JSONMap          `db:"raw_payload" json:"raw_payload"`
	NormalizedPayload db.JSONMap          `db:"normalized_payload" json:"normalized_payload"`
	ParseStatus       types.ParseStatus   `db:"parse_status" json:"parse_status"`
	ParseErrors       db.JSONMapSlice     `db:"parse_errors,omitempty" json:"parse_errors,omitempty"`
	OccurredAt        time.Time           `db:"occurred_at" json:"occurred_at"`
	IngestedAt        time.Time           `db:"ingested_at" json:"ingested_at"`
	CreatedAt         time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time           `db:"updated_at" json:"updated_at"`
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
}

func (s *SecurityEvent) ToResponse() *SecurityEventResponse {
	pid := ""
	if s.ParserID != nil {
		pid = s.ParserID.String()
	}
	var severity string
	if s.Severity != nil {
		severity = string(*s.Severity)
	} else {
		severity = "unknown"
	}
	return &SecurityEventResponse{
		ID:                s.ID.String(),
		SourceID:          s.SourceID.String(),
		Source:            s.Source,
		SourceEventID:     s.SourceEventID,
		EventType:         s.EventType,
		EventCategory:     string(s.EventCategory),
		Severity:          severity,
		ActorEmail:        s.ActorEmail,
		ActorUsername:     s.ActorUsername,
		IPAddress:         s.IPAddress,
		GeoCountry:        s.GeoCountry,
		GeoCity:           s.GeoCity,
		RawPayload:        s.RawPayload,
		NormalizedPayload: s.NormalizedPayload,
		ParserID:          &pid,
		ParseStatus:       string(s.ParseStatus),
		ParseErrors:       s.ParseErrors,
		OccurredAt:        s.OccurredAt,
		IngestedAt:        s.IngestedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

type CreateRawEventResponse struct {
	ID          uuid.UUID `db:"id"`
	CollectedAt time.Time `db:"collected_at"`
}

type RawEvent struct {
	ID             uuid.UUID `db:"id"`
	OrganizationID uuid.UUID `db:"organization_id"`
	SourceID       uuid.UUID `db:"source_id"`

	Provider  string `db:"provider"`
	EventType string `db:"event_type"`

	UserID    *string `db:"user_id,omitempty"`
	UserName  *string `db:"user_name,omitempty"`
	IPAddress *string `db:"ip_address,omitempty"`

	Application *string `db:"application,omitempty"`

	EventTimeStamp time.Time  `db:"event_timestamp"`
	CollectedAt    *time.Time `db:"collected_at"`

	Status string `db:"status"`

	ProviderStatus string `db:"provider_status"`

	LockedAt *time.Time `db:"locked_at,omitempty"`
	LockedBy *string    `db:"locked_by,omitempty"`

	ErrorMessage *string `db:"error_message,omitempty"`

	RawPayload db.JSONMap `db:"raw_payload"`
}
