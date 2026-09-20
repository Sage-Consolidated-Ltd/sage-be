package dto

import "time"

// GetSuggestedFixDiffQuery params
type GetSuggestedFixDiffRequest struct {
	SuggestionID string `json:"suggestion_id" query:"suggestion_id" validate:"required,uuid"`
	ParserID     string `json:"parser_id" query:"parser_id" validate:"required,uuid"`
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

type DataQualityScanResponse struct {
	ID                        string     `json:"id"`
	Status                    string     `json:"status"`
	QualityScore              *int       `json:"quality_score,omitempty"`
	ParsingErrors             *int64     `json:"parsing_errors,omitempty"`
	MissingFieldsPercentage   *float64   `json:"missing_fields_percentage,omitempty"`
	DuplicateEventsPercentage *float64   `json:"duplicate_events_percentage,omitempty"`
	UnmappedLogsCount         *int64     `json:"unmapped_logs_count,omitempty"`
	AISummary                 *string    `json:"ai_summary,omitempty"`
	StartedAt                 time.Time  `json:"started_at"`
	CompletedAt               *time.Time `json:"completed_at,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
}

type SuggestionResponse struct {
	ID             string                 `json:"id"`
	SourceID       *string                `json:"source_id,omitempty"`
	ParserID       *string                `json:"parser_id,omitempty"`
	Summary        string                 `json:"summary"`
	Recommendation string                 `json:"recommendation"`
	SuggestedFix   map[string]interface{} `json:"suggested_fix"`
	Confidence     float64                `json:"confidence"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	AppliedAt      *time.Time             `json:"applied_at,omitempty"`
}
