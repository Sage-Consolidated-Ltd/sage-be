package domain

import (
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

type SecurityEvent struct {
	ID                uuid.UUID           `json:"id" db:"id"`
	OrganizationID    uuid.UUID           `json:"organization_id" db:"organization_id"`
	SourceID          uuid.UUID           `json:"source_id" db:"source_id"`
	ParserID          *uuid.UUID          `json:"parser_id" db:"parser_id"`
	SourceEventID     *string             `json:"source_event_id" db:"source_event_id"`
	Source            string              `json:"source" db:"source"`
	EventType         string              `json:"event_type" db:"event_type"`
	EventCategory     types.EventCategory `json:"event_category" db:"event_category"`
	Severity          *types.Severity     `json:"severity" db:"severity"`
	ActorEmail        *string             `json:"actor_email" db:"actor_email"`
	ActorUsername     *string             `json:"actor_username" db:"actor_username"`
	IPAddress         *string             `json:"ip_address" db:"ip_address"`
	GeoCountry        *string             `json:"geo_country" db:"geo_country"`
	GeoCity           *string             `json:"geo_city" db:"geo_city"`
	RawPayload        db.JSONMap          `json:"raw_payload" db:"raw_payload"`
	NormalizedPayload db.JSONMap          `json:"normalized_payload" db:"normalized_payload"`
	ParseStatus       types.ParseStatus   `json:"parse_status" db:"parse_status"`
	ParseErrors       db.JSONMapSlice     `json:"parse_errors" db:"parse_errors"`
	SearchVector      *string             `json:"search_vector,omitempty" db:"search_vector"`
	OccurredAt        time.Time           `json:"occurred_at" db:"occurred_at"`
	IngestedAt        time.Time           `json:"ingested_at" db:"ingested_at"`
	CreatedAt         time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at" db:"updated_at"`
}

type RawEvent struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SourceID       *uuid.UUID `json:"source_id" db:"source_id"`
	UserID         *string    `json:"user_id" db:"user_id"`
	UserName       *string    `json:"user_name" db:"user_name"`
	Provider       string     `json:"provider" db:"provider"`
	EventType      string     `json:"event_type" db:"event_type"`
	IPAddress      *string    `json:"ip_address" db:"ip_address"`
	RawPayload     db.JSONMap `json:"raw_payload" db:"raw_payload"`
	EventTimeStamp time.Time  `json:"event_time_stamp" db:"event_time_stamp"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type CreateRawEventResponse struct {
	ID        uuid.UUID
	Payload   []byte
	CreatedAt time.Time
}

type ThreatsSummary struct {
	Critical       int64 `json:"critical" db:"critical"`
	High           int64 `json:"high" db:"high"`
	Medium         int64 `json:"medium" db:"medium"`
	Low            int64 `json:"low" db:"low"`
	NewInLast7Days int64 `json:"new_in_last_7_days" db:"new_in_last_7_days"`
	TotalThreats   int64 `json:"total_threats" db:"total_threats"`
}
