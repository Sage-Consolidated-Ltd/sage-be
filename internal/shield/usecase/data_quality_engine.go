package usecase

import (
	"context"
	"math"
	"sync"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/inbound"
)

// DataQualityEngine implements ports.inbound.DataQualityEngine for evaluating log event quality.
type DataQualityEngine struct {
	mu    sync.RWMutex
	rules []QualityRule
}

// NewDataQualityEngine initializes a DataQualityEngine pre-loaded with default Windows quality rules.
func NewDataQualityEngine() inbound.DataQualityEngine {
	engine := &DataQualityEngine{
		rules: make([]QualityRule, 0),
	}

	// Register default suite of Windows log quality rules
	engine.RegisterRule(&WinMissingRequiredFieldsRule{})
	engine.RegisterRule(&WinInvalidFieldsRule{})
	engine.RegisterRule(&WinMalformedValuesRule{})
	engine.RegisterRule(&WinInvalidTimestampRule{})
	engine.RegisterRule(&WinMissingMetadataRule{})
	engine.RegisterRule(&WinParseStatusRule{})
	engine.RegisterRule(NewWinDuplicateEventRule(10 * time.Minute))
	engine.RegisterRule(&WinPartialPopulationRule{})

	return engine
}

// RegisterRule appends a new QualityRule to the engine.
func (e *DataQualityEngine) RegisterRule(rule QualityRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// EvaluateEvent evaluates a single security event against all registered quality rules.
func (e *DataQualityEngine) EvaluateEvent(ctx context.Context, event *domain.SecurityEvent) (*domain.QualityResult, error) {
	if event == nil {
		return &domain.QualityResult{
			Score:       0.0,
			Status:      types.QualityError,
			EvaluatedAt: time.Now(),
			Issues: []domain.QualityIssue{
				{
					Code:     "NIL_EVENT",
					Message:  "Provided security event is nil",
					Severity: domain.QualitySeverityCritical,
					Category: "invalid_input",
				},
			},
		}, nil
	}

	e.mu.RLock()
	rules := make([]QualityRule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	var allIssues []domain.QualityIssue
	passedCount := 0

	for _, r := range rules {
		issues := r.Evaluate(event)
		if len(issues) == 0 {
			passedCount++
		} else {
			allIssues = append(allIssues, issues...)
		}
	}

	score := e.calculateScore(allIssues)
	status := e.determineStatus(score)

	eventIDStr := ""
	if event.SourceEventID != nil {
		eventIDStr = *event.SourceEventID
	} else if event.ID.String() != "" {
		eventIDStr = event.ID.String()
	}

	result := &domain.QualityResult{
		EventID:        eventIDStr,
		OrganizationID: event.OrganizationID.String(),
		Score:          score,
		Status:         status,
		TotalChecks:    len(rules),
		PassedChecks:   passedCount,
		Issues:         allIssues,
		EvaluatedAt:    time.Now(),
		Metadata: map[string]any{
			"rule_count": len(rules),
			"source":     event.Source,
		},
	}

	return result, nil
}

// EvaluateBatch evaluates a batch of events sequentially against quality rules.
func (e *DataQualityEngine) EvaluateBatch(ctx context.Context, events []*domain.SecurityEvent) ([]*domain.QualityResult, error) {
	results := make([]*domain.QualityResult, 0, len(events))
	for _, evt := range events {
		res, err := e.EvaluateEvent(ctx, evt)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

// calculateScore computes deterministic penalty-deducted score bounded in [0.0, 100.0].
func (e *DataQualityEngine) calculateScore(issues []domain.QualityIssue) float64 {
	score := 100.0
	for _, issue := range issues {
		score -= issue.Severity.Penalty()
	}
	if score < 0.0 {
		score = 0.0
	}
	return math.Round(score*100) / 100
}

// determineStatus maps calculated score to domain QualityStatus enum.
func (e *DataQualityEngine) determineStatus(score float64) types.QualityStatus {
	switch {
	case score >= 80.0:
		return types.QualityGood
	case score >= 60.0:
		return types.QualityWarning
	case score >= 40.0:
		return types.QualityPartial
	default:
		return types.QualityError
	}
}
