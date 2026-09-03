package domain

import (
	"encoding/json"
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
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	Name             string
	Description      *string
	Type             string
	Provider         *string
	Status           DataSourceStatus
	EventsToday      int64
	TotalEvents      int64
	LastEventAt      *time.Time
	LastSyncAt       *time.Time
	ErrorCount       int64
	DelayedByMinutes int
	Metadata         json.RawMessage
	LastCheckpoint   *string
	LastCheckpointAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type Checkpoint struct {
	LastCheckpoint   *string
	LastCheckpointAt *time.Time
}

type NormalizedEvent struct {
	ID          string
	Provider    string
	EventType   string
	UserID      string
	UserName    string
	IPAddress   string
	Timestamp   time.Time
	Raw         map[string]interface{}
	Application string
	Status      string
}
