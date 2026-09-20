package inbound

import (
	"context"
	"sage-backend/internal/shield/domain"
)

// IncidentEngine defines the inbound contract for evaluating security events against detection rules.
type IncidentEngine interface {
	// EvaluateEvent evaluates a security event against registered detection rules and returns created incidents.
	EvaluateEvent(ctx context.Context, event *domain.SecurityEvent) ([]*domain.Incident, error)

	// EvaluateBatch evaluates a batch of events sequentially or concurrently against detection rules.
	EvaluateBatch(ctx context.Context, events []*domain.SecurityEvent) ([]*domain.Incident, error)

	// RegisterRule registers a detection rule with the incident engine.
	RegisterRule(rule any) error

	// GetRegisteredRules returns metadata of all currently registered detection rules.
	GetRegisteredRules() []domain.RuleMetadata
}
