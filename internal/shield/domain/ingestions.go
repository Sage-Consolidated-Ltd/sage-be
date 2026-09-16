package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type IntegrationCredentials struct {
	ID             uuid.UUID
	IntegrationId  uuid.UUID
	Key            string
	EncryptedValue string
	ExpiresAt      sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type IntegrationStream struct {
	ID            uuid.UUID
	IntegrationId uuid.UUID
	StreamName    string
	LastOffset    *string
	LastEventAt   sql.NullTime
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type IngestionSession struct {
	ID            string
	IntegrationId string
	Status        string
	StartedAt     *time.Time
	EndedAt       *time.Time
	Error         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
