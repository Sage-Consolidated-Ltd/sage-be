package domain

import (
	"sage-backend/internal/shared/types"
	"time"

	"github.com/google/uuid"
)

type DataQualitySourceMetric struct {
	ID                      uuid.UUID
	OrganizationID          uuid.UUID
	ScanID                  uuid.UUID
	SourceID                uuid.UUID
	ParsingErrors           int64
	MissingFieldsPercentage float64
	UnmappedEvents          int64
	DuplicatePercentage     float64
	Status                  types.QualityStatus
	CreatedAt               time.Time
}
