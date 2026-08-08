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
	ID             uuid.UUID
	OrganizationID uuid.UUID
	SourceID       *uuid.UUID
	ParserID       *uuid.UUID
	Summary        string
	Recommendation string
	SuggestedFix   map[string]interface{}
	Confidence     float64
	Status         SuggestionStatus
	CreatedAt      time.Time
	AppliedAt      *time.Time
}
