package domain

import (
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

type SecurityEvent struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	SourceID          uuid.UUID
	ParserID          *uuid.UUID
	SourceEventID     *string
	Source            string
	EventType         string
	EventCategory     types.EventCategory
	Severity          *types.Severity
	ActorEmail        *string
	ActorUsername     *string
	IPAddress         *string
	GeoCountry        *string
	GeoCity           *string
	RawPayload        db.JSONMap
	NormalizedPayload db.JSONMap
	ParseStatus       types.ParseStatus
	ParseErrors       db.JSONMapSlice
	OccurredAt        time.Time
	IngestedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RawEvent struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	SourceID       *uuid.UUID
	UserID         *string
	UserName       *string
	Provider       string
	EventType      string
	IPAddress      *string
	RawPayload     db.JSONMap
	EventTimeStamp time.Time
	CreatedAt      time.Time
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
