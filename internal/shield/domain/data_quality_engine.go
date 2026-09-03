package domain

import (
	"time"

	"sage-backend/internal/shared/types"
)

// QualitySeverity indicates the severity level of a detected data quality issue.
type QualitySeverity string

const (
	QualitySeverityInfo     QualitySeverity = "info"
	QualitySeverityWarning  QualitySeverity = "warning"
	QualitySeverityError    QualitySeverity = "error"
	QualitySeverityCritical QualitySeverity = "critical"
)

func (s QualitySeverity) IsValid() bool {
	switch s {
	case QualitySeverityInfo, QualitySeverityWarning, QualitySeverityError, QualitySeverityCritical:
		return true
	default:
		return false
	}
}

// Penalty returns the score deduction associated with a quality issue severity.
func (s QualitySeverity) Penalty() float64 {
	switch s {
	case QualitySeverityCritical:
		return 25.0
	case QualitySeverityError:
		return 15.0
	case QualitySeverityWarning:
		return 5.0
	case QualitySeverityInfo:
		return 1.0
	default:
		return 0.0
	}
}

// QualityIssue represents a specific data quality anomaly or violation detected in an event.
type QualityIssue struct {
	Code           string                 `json:"code"`
	Message        string                 `json:"message"`
	Severity       QualitySeverity        `json:"severity"`
	Category       string                 `json:"category"`
	AffectedFields []string               `json:"affected_fields,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// QualityResult represents the complete data quality evaluation summary for an event.
type QualityResult struct {
	EventID        string              `json:"event_id"`
	OrganizationID string              `json:"organization_id"`
	Score          float64             `json:"score"`
	Status         types.QualityStatus `json:"status"`
	TotalChecks    int                 `json:"total_checks"`
	PassedChecks   int                 `json:"passed_checks"`
	Issues         []QualityIssue      `json:"issues"`
	EvaluatedAt    time.Time           `json:"evaluated_at"`
	Metadata       map[string]any      `json:"metadata,omitempty"`
}

// IsAcceptable returns true if the quality score meets the minimum usable threshold (>= 50).
func (q *QualityResult) IsAcceptable() bool {
	return q.Score >= 50.0
}
