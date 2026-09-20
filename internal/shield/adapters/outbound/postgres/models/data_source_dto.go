package models

import (
	"encoding/json"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type DataSourceDTO struct {
	ID               uuid.UUID `db:"id"`
	OrganizationID   uuid.UUID `db:"organization_id"`
	Name             string    `db:"name"`
	Description      *string   `db:"description,omitempty"`
	Type             string    `db:"type"`
	Provider         *string   `db:"provider,omitempty"`
	Status           string		`db:"status"`
	EventsToday      int64 `db:"events_today"`
	TotalEvents      int64 `db:"total_events"`
	LastEventAt      *time.Time `db:"last_event_at,omitempty"`
	LastSyncAt       *time.Time `db:"last_sync_at,omitempty"`
	ErrorCount       int64 `db:"error_count"`
	DelayedByMinutes int `db:"delayed_by_minutes"`
	Metadata         json.RawMessage `db:"metadata"`
	LastCheckpoint   *string `db:"last_checkpoint,omitempty"`
	LastCheckpointAt *time.Time `db:"last_checkpoint_at,omitempty"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at,omitempty"`
}

func (dto *DataSourceDTO) ToDomain() *domain.DataSource {
	return &domain.DataSource{
		ID:               dto.ID,
		OrganizationID:   dto.OrganizationID,
		Name:             dto.Name,
		Description:      dto.Description,
		Type:             dto.Type,
		Provider:         dto.Provider,
		Status:           domain.DataSourceStatus(dto.Status),
		EventsToday:      dto.EventsToday,
		TotalEvents:      dto.TotalEvents,
		LastEventAt:      dto.LastEventAt,
		LastSyncAt:       dto.LastSyncAt,
		ErrorCount:       dto.ErrorCount,
		DelayedByMinutes: dto.DelayedByMinutes,
		Metadata:         dto.Metadata,
		LastCheckpoint:   dto.LastCheckpoint,
		LastCheckpointAt: dto.LastCheckpointAt,
		CreatedAt:        dto.CreatedAt,
		UpdatedAt:        dto.UpdatedAt,
		DeletedAt:        dto.DeletedAt,
	}
}