package usecase

import (
	"context"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
)

// DetectionRule is the base interface that all detection rules must implement.
type DetectionRule interface {
	Metadata() domain.RuleMetadata
}

// StatelessRule represents a detection rule evaluated against a single security event.
type StatelessRule interface {
	DetectionRule
	Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error)
}

// StatefulRule represents a multi-event correlation rule requiring a sliding-window state store.
type StatefulRule interface {
	DetectionRule
	EvaluateStateful(ctx context.Context, event *domain.SecurityEvent, store outbound.CorrelationStore) (*domain.DetectionResult, error)
}
