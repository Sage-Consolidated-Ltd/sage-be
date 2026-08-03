package domain

import (
	"time"

	"github.com/google/uuid"
)

type SuggestionStatus string

const (
	SuggestionStatusPending   SuggestionStatus = "pending"
	SuggestionStatusApplied   SuggestionStatus = "applied"
	SuggestionStatusDismissed SuggestionStatus = "dismissed"
)

func (s SuggestionStatus) IsValid() bool {
	switch s {
	case SuggestionStatusPending, SuggestionStatusApplied, SuggestionStatusDismissed:
		return true
	default:
		return false
	}
}

type DataQualitySuggestion struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
	SourceID       *uuid.UUID             `db:"source_id,omitempty" json:"source_id,omitempty"`
	ParserID       *uuid.UUID             `db:"parser_id,omitempty" json:"parser_id,omitempty"`
	Summary        string                 `db:"summary" json:"summary"`
	Recommendation string                 `db:"recommendation" json:"recommendation"`
	SuggestedFix   map[string]interface{} `db:"suggested_fix" json:"suggested_fix"`
	Confidence     float64                `db:"confidence" json:"confidence"`
	Status         SuggestionStatus       `db:"status" json:"status"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
	AppliedAt      *time.Time             `db:"applied_at,omitempty" json:"applied_at,omitempty"`
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

func (d *DataQualitySuggestion) ToResponse() *SuggestionResponse {
	sid := ""
	if d.SourceID != nil {
		sid = d.SourceID.String()
	}
	pid := ""
	if d.ParserID != nil {
		pid = d.ParserID.String()
	}
	return &SuggestionResponse{
		ID:             d.ID.String(),
		SourceID:       &sid,
		ParserID:       &pid,
		Summary:        d.Summary,
		Recommendation: d.Recommendation,
		SuggestedFix:   d.SuggestedFix,
		Confidence:     d.Confidence,
		Status:         string(d.Status),
		CreatedAt:      d.CreatedAt,
		AppliedAt:      d.AppliedAt,
	}
}
