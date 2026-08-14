package inbound

import (
	"context"
	"sage-backend/internal/shield/domain"
)

// DataQualityEngine defines the inbound contract for evaluating data quality on security log events.
type DataQualityEngine interface {
	// EvaluateEvent evaluates a single security event against all enabled data quality rules.
	EvaluateEvent(ctx context.Context, event *domain.SecurityEvent) (*domain.QualityResult, error)

	// EvaluateBatch evaluates a slice of security events against all enabled data quality rules.
	EvaluateBatch(ctx context.Context, events []*domain.SecurityEvent) ([]*domain.QualityResult, error)
}
