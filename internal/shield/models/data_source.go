package models

import (
	"time"

	"github.com/google/uuid"
)

type DataSourceStatus string

const (
	DataSourceStatusActive       DataSourceStatus = "active"
	DataSourceStatusDelayed      DataSourceStatus = "delayed"
	DataSourceStatusError        DataSourceStatus = "error"
	DataSourceStatusDisabled     DataSourceStatus = "disabled"
	DataSourceStatusDisconnected DataSourceStatus = "disconnected"
)

func (d DataSourceStatus) IsValid() bool {
	switch d {
	case DataSourceStatusActive, DataSourceStatusDelayed, DataSourceStatusError,
		DataSourceStatusDisabled, DataSourceStatusDisconnected:
		return true
	default:
		return false
	}
}

type DataSource struct {
	ID               uuid.UUID              `db:"id" json:"id"`
	OrganizationID   uuid.UUID              `db:"organization_id" json:"organization_id"`
	Name             string                 `db:"name" json:"name"`
	Description      *string                `db:"description,omitempty" json:"description,omitempty"`
	Type             string                 `db:"type" json:"type"`
	Provider         *string                `db:"provider,omitempty" json:"provider,omitempty"`
	Status           DataSourceStatus       `db:"status" json:"status"`
	EventsToday      int64                  `db:"events_today" json:"events_today"`
	TotalEvents      int64                  `db:"total_events" json:"total_events"`
	LastEventAt      *time.Time             `db:"last_event_at,omitempty" json:"last_event_at,omitempty"`
	LastSyncAt       *time.Time             `db:"last_sync_at,omitempty" json:"last_sync_at,omitempty"`
	ErrorCount       int64                  `db:"error_count" json:"error_count"`
	DelayedByMinutes int                    `db:"delayed_by_minutes" json:"delayed_by_minutes"`
	Metadata         map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt        time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `db:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time             `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type DataSourceResponse struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      *string                `json:"description,omitempty"`
	Type             string                 `json:"type"`
	Provider         *string                `json:"provider,omitempty"`
	Status           string                 `json:"status"`
	EventsToday      int64                  `json:"events_today"`
	TotalEvents      int64                  `json:"total_events"`
	LastEventAt      *time.Time             `json:"last_event_at,omitempty"`
	LastSyncAt       *time.Time             `json:"last_sync_at,omitempty"`
	ErrorCount       int64                  `json:"error_count"`
	DelayedByMinutes int                    `json:"delayed_by_minutes"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

func (d *DataSource) ToResponse() *DataSourceResponse {
	meta := d.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	}
	return &DataSourceResponse{
		ID:               d.ID.String(),
		Name:             d.Name,
		Description:      d.Description,
		Type:             d.Type,
		Provider:         d.Provider,
		Status:           string(d.Status),
		EventsToday:      d.EventsToday,
		TotalEvents:      d.TotalEvents,
		LastEventAt:      d.LastEventAt,
		LastSyncAt:       d.LastSyncAt,
		ErrorCount:       d.ErrorCount,
		DelayedByMinutes: d.DelayedByMinutes,
		Metadata:         meta,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}
