package models

import (
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type SecurityEventDTO struct {
	ID                uuid.UUID           `db:"id"`
	OrganizationID    uuid.UUID           `db:"organization_id"`
	SourceID          uuid.UUID           `db:"source_id"`
	ParserID          *uuid.UUID          `db:"parser_id"`
	SourceEventID     *string             `db:"source_event_id"`
	Source            string              `db:"source"`
	EventType         string              `db:"event_type"`
	EventCategory     types.EventCategory `db:"event_category"`
	Severity          *types.Severity     `db:"severity"`
	ActorEmail        *string             `db:"actor_email"`
	ActorUsername     *string             `db:"actor_username"`
	IPAddress         *string             `db:"ip_address"`
	GeoCountry        *string             `db:"geo_country"`
	GeoCity           *string             `db:"geo_city"`
	RawPayload        db.JSONMap          `db:"raw_payload"`
	NormalizedPayload db.JSONMap          `db:"normalized_payload"`
	ParseStatus       types.ParseStatus   `db:"parse_status"`
	ParseErrors       db.JSONMapSlice     `db:"parse_errors"`
	SearchVector      *string             `db:"search_vector"`
	OccurredAt        time.Time           `db:"occurred_at"`
	IngestedAt        time.Time           `db:"ingested_at"`
	CreatedAt         time.Time           `db:"created_at"`
	UpdatedAt         time.Time           `db:"updated_at"`
}

func (dto *SecurityEventDTO) ToDomain() *domain.SecurityEvent {
	if dto == nil {
		return nil
	}
	return &domain.SecurityEvent{
		ID:                dto.ID,
		OrganizationID:    dto.OrganizationID,
		SourceID:          dto.SourceID,
		ParserID:          dto.ParserID,
		SourceEventID:     dto.SourceEventID,
		Source:            dto.Source,
		EventType:         dto.EventType,
		EventCategory:     dto.EventCategory,
		Severity:          dto.Severity,
		ActorEmail:        dto.ActorEmail,
		ActorUsername:     dto.ActorUsername,
		IPAddress:         dto.IPAddress,
		GeoCountry:        dto.GeoCountry,
		GeoCity:           dto.GeoCity,
		RawPayload:        dto.RawPayload,
		NormalizedPayload: dto.NormalizedPayload,
		ParseStatus:       dto.ParseStatus,
		ParseErrors:       dto.ParseErrors,
		SearchVector:      dto.SearchVector,
		OccurredAt:        dto.OccurredAt,
		IngestedAt:        dto.IngestedAt,
		CreatedAt:         dto.CreatedAt,
		UpdatedAt:         dto.UpdatedAt,
	}
}

type RawEventDTO struct {
	ID             uuid.UUID  `db:"id"`
	OrganizationID uuid.UUID  `db:"organization_id"`
	SourceID       *uuid.UUID `db:"source_id"`
	UserID         *string    `db:"user_id"`
	UserName       *string    `db:"user_name"`
	Provider       string     `db:"provider"`
	EventType      string     `db:"event_type"`
	IPAddress      *string    `db:"ip_address"`
	RawPayload     db.JSONMap `db:"raw_payload"`
	EventTimeStamp time.Time  `db:"event_time_stamp"`
	CreatedAt      time.Time  `db:"created_at"`
}

func (dto *RawEventDTO) ToDomain() *domain.RawEvent {
	if dto == nil {
		return nil
	}
	return &domain.RawEvent{
		ID:             dto.ID,
		OrganizationID: dto.OrganizationID,
		SourceID:       dto.SourceID,
		UserID:         dto.UserID,
		UserName:       dto.UserName,
		Provider:       dto.Provider,
		EventType:      dto.EventType,
		IPAddress:      dto.IPAddress,
		RawPayload:     dto.RawPayload,
		EventTimeStamp: dto.EventTimeStamp,
		CreatedAt:      dto.CreatedAt,
	}
}

type ThreatsSummaryDTO struct {
	Critical       int64 `db:"critical"`
	High           int64 `db:"high"`
	Medium         int64 `db:"medium"`
	Low            int64 `db:"low"`
	NewInLast7Days int64 `db:"new_in_last_7_days"`
	TotalThreats   int64 `db:"total_threats"`
}

func (dto *ThreatsSummaryDTO) ToDomain() *domain.ThreatsSummary {
	if dto == nil {
		return nil
	}
	return &domain.ThreatsSummary{
		Critical:       dto.Critical,
		High:           dto.High,
		Medium:         dto.Medium,
		Low:            dto.Low,
		NewInLast7Days: dto.NewInLast7Days,
		TotalThreats:   dto.TotalThreats,
	}
}
