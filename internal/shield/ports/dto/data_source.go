package dto

import (
	"encoding/json"
	"time"
)

type DataSourceResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	Type             string          `json:"type"`
	Provider         *string         `json:"provider,omitempty"`
	Status           string          `json:"status"`
	EventsToday      int64           `json:"events_today"`
	TotalEvents      int64           `json:"total_events"`
	LastEventAt      *time.Time      `json:"last_event_at,omitempty"`
	LastSyncAt       *time.Time      `json:"last_sync_at,omitempty"`
	ErrorCount       int64           `json:"error_count"`
	DelayedByMinutes int             `json:"delayed_by_minutes"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	LastCheckpoint   *string         `json:"last_checkpoint,omitempty"`
	LastCheckpointAt *time.Time      `json:"last_checkpoint_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type NormalizedEvent struct {
	ID          string                 `json:"id"`
	Provider    string                 `json:"provider"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	UserName    string                 `json:"user_name"`
	IPAddress   string                 `json:"ip_address"`
	Timestamp   time.Time              `json:"timestamp"`
	Raw         map[string]interface{} `json:"raw"`
	Application string                 `json:"application,omitempty"`
	Status      string                 `json:"status,omitempty"`
}
