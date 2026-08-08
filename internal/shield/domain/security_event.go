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
