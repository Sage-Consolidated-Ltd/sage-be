package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

// IncidentEngine implements ports.inbound.IncidentEngine for security event detection.
type IncidentEngine struct {
	mu               sync.RWMutex
	rules            []DetectionRule
	correlationStore outbound.CorrelationStore
}

// NewIncidentEngine initializes an IncidentEngine instance with a correlation state store.
func NewIncidentEngine(store outbound.CorrelationStore) inbound.IncidentEngine {
	return &IncidentEngine{
		rules:            make([]DetectionRule, 0),
		correlationStore: store,
	}
}

// RegisterRule registers a DetectionRule (stateless or stateful) with the engine.
func (e *IncidentEngine) RegisterRule(rule any) error {
	if rule == nil {
		return fmt.Errorf("cannot register nil detection rule")
	}

	detRule, ok := rule.(DetectionRule)
	if !ok {
		return fmt.Errorf("provided rule does not implement DetectionRule interface")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	meta := detRule.Metadata()
	for _, existing := range e.rules {
		if existing.Metadata().ID == meta.ID {
			return fmt.Errorf("rule with ID '%s' is already registered", meta.ID)
		}
	}

	e.rules = append(e.rules, detRule)
	return nil
}

// GetRegisteredRules returns metadata of all currently registered detection rules.
func (e *IncidentEngine) GetRegisteredRules() []domain.RuleMetadata {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]domain.RuleMetadata, 0, len(e.rules))
	for _, r := range e.rules {
		result = append(result, r.Metadata())
	}
	return result
}

// EvaluateEvent evaluates a security event against all enabled detection rules.
func (e *IncidentEngine) EvaluateEvent(ctx context.Context, event *domain.SecurityEvent) ([]*domain.Incident, error) {
	if event == nil {
		return []*domain.Incident{}, nil
	}

	e.mu.RLock()
	rules := make([]DetectionRule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	var incidents []*domain.Incident

	for _, rule := range rules {
		meta := rule.Metadata()
		if !meta.Enabled {
			continue
		}

		var detResult *domain.DetectionResult
		var err error

		if meta.IsStateful {
			statefulRule, ok := rule.(StatefulRule)
			if !ok {
				continue
			}
			detResult, err = statefulRule.EvaluateStateful(ctx, event, e.correlationStore)
		} else {
			statelessRule, ok := rule.(StatelessRule)
			if !ok {
				continue
			}
			detResult, err = statelessRule.Evaluate(event)
		}

		if err != nil {
			// Log error or continue to evaluate other rules
			continue
		}

		if detResult != nil && detResult.Matched {
			inc := e.buildIncident(event, meta, detResult)
			incidents = append(incidents, inc)
		}
	}

	return incidents, nil
}

// EvaluateBatch evaluates a slice of events against registered detection rules.
func (e *IncidentEngine) EvaluateBatch(ctx context.Context, events []*domain.SecurityEvent) ([]*domain.Incident, error) {
	var allIncidents []*domain.Incident
	for _, event := range events {
		incidents, err := e.EvaluateEvent(ctx, event)
		if err != nil {
			return nil, err
		}
		if len(incidents) > 0 {
			allIncidents = append(allIncidents, incidents...)
		}
	}
	return allIncidents, nil
}

// buildIncident constructs a full domain.Incident entity from a detection match.
func (e *IncidentEngine) buildIncident(
	event *domain.SecurityEvent,
	meta domain.RuleMetadata,
	detResult *domain.DetectionResult,
) *domain.Incident {
	now := time.Now()
	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}

	title := fmt.Sprintf("[%s] %s Detected", meta.Severity, meta.Name)
	summary := meta.Description

	// Ensure trigger event ID and evidence fields are populated
	evidence := detResult.Evidence
	evidence.RuleID = meta.ID
	evidence.RuleName = meta.Name
	evidence.Severity = detResult.Severity
	evidence.Description = meta.Description

	if evidence.TriggerEventID == "" {
		if event.SourceEventID != nil {
			evidence.TriggerEventID = *event.SourceEventID
		} else {
			evidence.TriggerEventID = event.ID.String()
		}
	}

	if evidence.TargetUser == "" && event.ActorUsername != nil {
		evidence.TargetUser = *event.ActorUsername
	}
	if evidence.SourceIP == "" && event.IPAddress != nil {
		evidence.SourceIP = *event.IPAddress
	}

	return &domain.Incident{
		ID:             uuid.New(),
		OrganizationID: event.OrganizationID,
		RuleID:         meta.ID,
		RuleName:       meta.Name,
		Category:       meta.Category,
		Severity:       detResult.Severity,
		Status:         domain.IncidentStatusNew,
		Title:          title,
		Summary:        summary,
		Evidence:       evidence,
		OccurredAt:     occurredAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
