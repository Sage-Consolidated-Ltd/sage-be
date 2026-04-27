package models

import (
	"sage-backend/internal/shared/types"
	"time"

	"github.com/google/uuid"
)

type DataQualitySourceMetric struct {
	ID                      uuid.UUID           `db:"id" json:"id"`
	OrganizationID          uuid.UUID           `db:"organization_id" json:"organization_id"`
	ScanID                  uuid.UUID           `db:"scan_id" json:"scan_id"`
	SourceID                uuid.UUID           `db:"source_id" json:"source_id"`
	ParsingErrors           int64               `db:"parsing_errors" json:"parsing_errors"`
	MissingFieldsPercentage float64             `db:"missing_fields_percentage" json:"missing_fields_percentage"`
	UnmappedEvents          int64               `db:"unmapped_events" json:"unmapped_events"`
	DuplicatePercentage     float64             `db:"duplicate_percentage" json:"duplicate_percentage"`
	Status                  types.QualityStatus `db:"status" json:"status"`
	CreatedAt               time.Time           `db:"created_at" json:"created_at"`
}

type DataQualityBreakdownResponse struct {
	SourceID                string  `json:"source_id"`
	SourceName              string  `json:"source_name"`
	ParsingErrors           int64   `json:"parsing_errors"`
	MissingFieldsPercentage float64 `json:"missing_fields_percentage"`
	UnmappedEvents          int64   `json:"unmapped_events"`
	DuplicatePercentage     float64 `json:"duplicate_percentage"`
	Status                  string  `json:"status"`
}

func (d *DataQualitySourceMetric) ToResponse(sourceName string) *DataQualityBreakdownResponse {
	return &DataQualityBreakdownResponse{
		SourceID:                d.SourceID.String(),
		SourceName:              sourceName,
		ParsingErrors:           d.ParsingErrors,
		MissingFieldsPercentage: d.MissingFieldsPercentage,
		UnmappedEvents:          d.UnmappedEvents,
		DuplicatePercentage:     d.DuplicatePercentage,
		Status:                  string(d.Status),
	}
}
