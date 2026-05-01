package models

import (
	"time"

	"github.com/google/uuid"
)

type DataQualityScan struct {
	ID                     uuid.UUID `db:"id" json:"id"`
	OrganizationID         uuid.UUID `db:"organization_id" json:"organization_id"`
	Status                 string    `db:"status" json:"status"` // running, completed, failed
	QualityScore           *int      `db:"quality_score" json:"quality_score"`
	ParsingErrors          *int64    `db:"parsing_errors" json:"parsing_errors"`
	MissingFieldsPercentage *float64 `db:"missing_fields_percentage" json:"missing_fields_percentage"`
	DuplicateEventsPercentage *float64 `db:"duplicate_events_percentage" json:"duplicate_events_percentage"`
	UnmappedLogsCount      *int64    `db:"unmapped_logs_count" json:"unmapped_logs_count"`
	AISummary              *string   `db:"ai_summary" json:"ai_summary,omitempty"`
	StartedAt             time.Time `db:"started_at" json:"started_at"`
	CompletedAt           *time.Time `db:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}

type DataQualityScanResponse struct {
	ID                     string     `json:"id"`
	Status                 string     `json:"status"`
	QualityScore           *int       `json:"quality_score,omitempty"`
	ParsingErrors          *int64     `json:"parsing_errors,omitempty"`
	MissingFieldsPercentage *float64  `json:"missing_fields_percentage,omitempty"`
	DuplicateEventsPercentage *float64 `json:"duplicate_events_percentage,omitempty"`
	UnmappedLogsCount      *int64     `json:"unmapped_logs_count,omitempty"`
	AISummary              *string    `json:"ai_summary,omitempty"`
	StartedAt              time.Time  `json:"started_at"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}

func (d *DataQualityScan) ToResponse() *DataQualityScanResponse {
	return &DataQualityScanResponse{
		ID:                     d.ID.String(),
		Status:                 d.Status,
		QualityScore:           d.QualityScore,
		ParsingErrors:          d.ParsingErrors,
		MissingFieldsPercentage: d.MissingFieldsPercentage,
		DuplicateEventsPercentage: d.DuplicateEventsPercentage,
		UnmappedLogsCount:      d.UnmappedLogsCount,
		AISummary:              d.AISummary,
		StartedAt:             d.StartedAt,
		CompletedAt:           d.CompletedAt,
		CreatedAt:             d.CreatedAt,
	}
}
