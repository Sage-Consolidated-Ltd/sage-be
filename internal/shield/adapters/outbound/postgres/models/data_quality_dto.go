package models

import (
	"encoding/json"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type DataQualityScanDTO struct {
	ID                        uuid.UUID  `db:"id"`
	OrganizationID            uuid.UUID  `db:"organization_id"`
	Status                    string     `db:"status"`
	QualityScore              *float64   `db:"quality_score"`
	ParsingErrors             *int64     `db:"parsing_errors"`
	MissingFieldsPercentage   *float64   `db:"missing_fields_percentage"`
	DuplicateEventsPercentage *float64   `db:"duplicate_events_percentage"`
	UnmappedLogsCount         *int64     `db:"unmapped_logs_count"`
	AISummary                 *string    `db:"ai_summary"`
	StartedAt                 time.Time  `db:"started_at"`
	CompletedAt               *time.Time `db:"completed_at"`
	CreatedAt                 time.Time  `db:"created_at"`
}

func (dto *DataQualityScanDTO) ToDomain() *domain.DataQualityScan {
	if dto == nil {
		return nil
	}

	var score *int
	if dto.QualityScore != nil {
		s := int(*dto.QualityScore)
		score = &s
	}

	return &domain.DataQualityScan{
		ID:                        dto.ID,
		OrganizationID:            dto.OrganizationID,
		Status:                    dto.Status,
		QualityScore:              score,
		ParsingErrors:             dto.ParsingErrors,
		MissingFieldsPercentage:   dto.MissingFieldsPercentage,
		DuplicateEventsPercentage: dto.DuplicateEventsPercentage,
		UnmappedLogsCount:         dto.UnmappedLogsCount,
		AISummary:                 dto.AISummary,
		StartedAt:                 dto.StartedAt,
		CompletedAt:               dto.CompletedAt,
		CreatedAt:                 dto.CreatedAt,
	}
}

type DataQualitySourceMetricDTO struct {
	ID                      uuid.UUID           `db:"id"`
	OrganizationID          uuid.UUID           `db:"organization_id"`
	ScanID                  uuid.UUID           `db:"scan_id"`
	SourceID                uuid.UUID           `db:"source_id"`
	ParsingErrors           int64               `db:"parsing_errors"`
	MissingFieldsPercentage float64             `db:"missing_fields_percentage"`
	UnmappedEvents          int64               `db:"unmapped_events"`
	DuplicatePercentage     float64             `db:"duplicate_percentage"`
	Status                  types.QualityStatus `db:"status"`
	CreatedAt               time.Time           `db:"created_at"`
}

func (dto *DataQualitySourceMetricDTO) ToDomain() *domain.DataQualitySourceMetric {
	if dto == nil {
		return nil
	}
	return &domain.DataQualitySourceMetric{
		ID:                      dto.ID,
		OrganizationID:          dto.OrganizationID,
		ScanID:                  dto.ScanID,
		SourceID:                dto.SourceID,
		ParsingErrors:           dto.ParsingErrors,
		MissingFieldsPercentage: dto.MissingFieldsPercentage,
		UnmappedEvents:          dto.UnmappedEvents,
		DuplicatePercentage:     dto.DuplicatePercentage,
		Status:                  dto.Status,
		CreatedAt:               dto.CreatedAt,
	}
}

type DataQualitySuggestionDTO struct {
	ID             uuid.UUID               `db:"id"`
	OrganizationID uuid.UUID               `db:"organization_id"`
	SourceID       *uuid.UUID              `db:"source_id"`
	ParserID       *uuid.UUID              `db:"parser_id"`
	Summary        string                  `db:"summary"`
	Recommendation string                  `db:"recommendation"`
	SuggestedFix   json.RawMessage         `db:"suggested_fix"`
	Confidence     float64                 `db:"confidence"`
	Status         domain.SuggestionStatus `db:"status"`
	CreatedAt      time.Time               `db:"created_at"`
	AppliedAt      *time.Time              `db:"applied_at"`
}

func (dto *DataQualitySuggestionDTO) ToDomain() *domain.DataQualitySuggestion {
	if dto == nil {
		return nil
	}

	var fix map[string]interface{}
	if len(dto.SuggestedFix) > 0 {
		_ = json.Unmarshal(dto.SuggestedFix, &fix)
	}

	return &domain.DataQualitySuggestion{
		ID:             dto.ID,
		OrganizationID: dto.OrganizationID,
		SourceID:       dto.SourceID,
		ParserID:       dto.ParserID,
		Summary:        dto.Summary,
		Recommendation: dto.Recommendation,
		SuggestedFix:   fix,
		Confidence:     dto.Confidence,
		Status:         dto.Status,
		CreatedAt:      dto.CreatedAt,
		AppliedAt:      dto.AppliedAt,
	}
}
