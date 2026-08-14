package outbound

import (
	"context"
	"time"

	"sage-backend/internal/shield/domain"
)

// CorrelationStore defines the outbound contract for stateful event windowing and correlation.
type CorrelationStore interface {
	// AddEvent adds an event to the state store under the specified correlation key.
	AddEvent(ctx context.Context, key string, event *domain.SecurityEvent, ttl time.Duration) error

	// GetEvents retrieves all stored events for a key within the specified time window.
	GetEvents(ctx context.Context, key string, window time.Duration) ([]*domain.SecurityEvent, error)

	// GetCount returns the number of active events stored under key within window duration.
	GetCount(ctx context.Context, key string, window time.Duration) (int, error)

	// Clear removes all stored state associated with the specified correlation key.
	Clear(ctx context.Context, key string) error
}
