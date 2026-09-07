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
	RuleCategoryInitialAccess  RuleCategory = "initial_access"
	RuleCategoryExecution      RuleCategory = "execution"
	RuleCategoryPersistence    RuleCategory = "persistence"
	RuleCategoryDefenseEvasion RuleCategory = "defense_evasion"
	RuleCategoryCredentialAccess RuleCategory = "credential_access"
	RuleCategoryDiscovery      RuleCategory = "discovery"
	RuleCategoryLateralMovement RuleCategory = "lateral_movement"
	RuleCategoryCollection     RuleCategory = "collection"
	RuleCategoryExfiltration   RuleCategory = "exfiltration"
	RuleCategoryImpact         RuleCategory = "impact"
	RuleCategoryMetaBreach     RuleCategory = "meta_breach"
)

// Alert represents an atomic detected threat signature before correlation.
type Alert struct {
	ID            uuid.UUID              `json:"id"`
	ThreatLabel   string                 `json:"threat_label"`
	MITRE         string                 `json:"mitre"`
	LogSource     string                 `json:"log_source"` // Security, Sysmon, System, Application
	EventID       string                 `json:"event_id"`
	EntityHost    string                 `json:"entity_host,omitempty"`
	EntityAccount string                 `json:"entity_account,omitempty"`
	EntityIP      string                 `json:"entity_ip,omitempty"`
	RawEvent      *SecurityEvent         `json:"raw_event,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	DetectedAt    time.Time              `json:"detected_at"`
}

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
	Score          int            `json:"score"`
	Priority       string         `json:"priority"`
	EntityKey      string         `json:"entity_key"`
	Status         IncidentStatus `json:"status"`
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	Evidence       Evidence       `json:"evidence"`
	OccurredAt     time.Time      `json:"occurred_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
