package domain

import (
	"time"

	"github.com/google/uuid"
)

type DataQualityScan struct {
	ID                        uuid.UUID
	OrganizationID            uuid.UUID
	Status                    string // running, completed, failed
	QualityScore              *int
	ParsingErrors             *int64
	MissingFieldsPercentage   *float64
	DuplicateEventsPercentage *float64
	UnmappedLogsCount         *int64
	AISummary                 *string
	StartedAt                 time.Time
	CompletedAt               *time.Time
	CreatedAt                 time.Time
}
