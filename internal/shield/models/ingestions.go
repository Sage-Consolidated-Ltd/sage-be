package models

import (
	"database/sql"
	"time"
)

type Integration struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	TenantId       string    `json:"tenant_id" db:"tenant_id"`
	Provider       string    `json:"provider" db:"provider"`
	ConnectionType string    `json:"connection_type" db:"connection_type"`
	Status         string    `json:"status" db:"status"`
	Config         []byte    `json:"config" db:"config"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type IntegrationCredentials struct {
	ID             string       `json:"id" db:"id"`
	IntegrationId  string       `json:"integration_id" db:"integration_id"`
	Key            string       `json:"key" db:"key"`
	EncryptedValue string       `json:"encrypted_value" db:"encrypted_value"`
	ExpiresAt      sql.NullTime `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

type IntegrationStream struct {
	ID            string       `json:"id" db:"id"`
	IntegrationId string       `json:"integration_id" db:"integration_id"`
	StreamName    string       `json:"stream_name" db:"stream_name"`
	LastOffset    *string      `json:"last_offset" db:"last_offset"`
	LastEventAt   sql.NullTime `json:"last_event_at" db:"last_event_at"`
	Status        string       `json:"status" db:"status"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

type IngestionSession struct {
	ID            string     `json:"id" db:"id"`
	IntegrationId string     `json:"integration_id" db:"integration_id"`
	Status        string     `json:"status" db:"status"`
	StartedAt     *time.Time `json:"started_at" db:"started_at"`
	EndedAt       *time.Time `json:"ended_at" db:"ended_at"`
	Error         *string    `json:"error" db:"error"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type IntegrationResponse struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Provider       string                 `json:"provider"`
	ConnectionType string                 `json:"connection_type"`
	Status         string                 `json:"status"`
	Config         map[string]interface{} `json:"config"`
	CreatedAt      time.Time              `json:"created_at"`
}
