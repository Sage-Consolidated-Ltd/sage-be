package domain

import (
	"time"

	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

// IncidentStatus represents the operational lifecycle state of a security incident.
type IncidentStatus string

const (
	IncidentStatusNew           IncidentStatus = "new"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusContained     IncidentStatus = "contained"
	IncidentStatusResolved      IncidentStatus = "resolved"
	IncidentStatusDismissed     IncidentStatus = "dismissed"
	IncidentStatusNeedsReview   IncidentStatus = "needs_review"
)

func (s IncidentStatus) IsValid() bool {
	switch s {
	case IncidentStatusNew, IncidentStatusInvestigating, IncidentStatusContained,
		IncidentStatusResolved, IncidentStatusDismissed, IncidentStatusNeedsReview:
		return true
	default:
		return false
	}
}

// RuleCategory classifies security detection rules into standard functional categories.
type RuleCategory string

const (
	RuleCategoryAuthentication RuleCategory = "authentication"
	RuleCategoryAccount        RuleCategory = "account_activity"
	RuleCategoryPrivilege      RuleCategory = "privilege_security"
	RuleCategoryProcessService RuleCategory = "process_service"
)

// RuleMetadata holds declarative information about a detection rule.
type RuleMetadata struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Category       RuleCategory   `json:"category"`
	Severity       types.Severity `json:"severity"`
	Enabled        bool           `json:"enabled"`
	IsStateful     bool           `json:"is_stateful"`
	WindowDuration time.Duration  `json:"window_duration,omitempty"`
	ThresholdCount int            `json:"threshold_count,omitempty"`
}

// DetectionResult is produced when an event matches a detection rule.
type DetectionResult struct {
	RuleMetadata RuleMetadata           `json:"rule_metadata"`
	Matched      bool                   `json:"matched"`
	Severity     types.Severity         `json:"severity"`
	Evidence     Evidence               `json:"evidence"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// Incident represents a verified security detection requiring investigation or response.
type Incident struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	RuleID         string         `json:"rule_id"`
	RuleName       string         `json:"rule_name"`
	Category       RuleCategory   `json:"category"`
	Severity       types.Severity `json:"severity"`
	Status         IncidentStatus `json:"status"`
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	Evidence       Evidence       `json:"evidence"`
	OccurredAt     time.Time      `json:"occurred_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
