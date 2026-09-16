package memory

import (
	"context"
	"sync"
	"time"

	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
)

type storedEvent struct {
	event     *domain.SecurityEvent
	timestamp time.Time
}

// CorrelationStore implements ports.outbound.CorrelationStore with a thread-safe in-memory sliding window.
type CorrelationStore struct {
	mu    sync.RWMutex
	store map[string][]storedEvent
}

// NewCorrelationStore initializes an in-memory CorrelationStore.
func NewCorrelationStore() outbound.CorrelationStore {
	return &CorrelationStore{
		store: make(map[string][]storedEvent),
	}
}

// AddEvent adds an event to the state store under the given key and prunes events older than ttl.
func (c *CorrelationStore) AddEvent(ctx context.Context, key string, event *domain.SecurityEvent, ttl time.Duration) error {
	if event == nil || key == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-ttl)

	// Filter existing events for key to drop expired entries
	existing := c.store[key]
	valid := make([]storedEvent, 0, len(existing)+1)

	for _, se := range existing {
		if se.timestamp.After(cutoff) {
			valid = append(valid, se)
		}
	}

	eventTime := event.OccurredAt
	if eventTime.IsZero() {
		eventTime = now
	}

	valid = append(valid, storedEvent{
		event:     event,
		timestamp: eventTime,
	})

	c.store[key] = valid
	return nil
}

// GetEvents retrieves all stored events for key within the specified window duration.
func (c *CorrelationStore) GetEvents(ctx context.Context, key string, window time.Duration) ([]*domain.SecurityEvent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-window)

	existing, exists := c.store[key]
	if !exists {
		return []*domain.SecurityEvent{}, nil
	}

	result := make([]*domain.SecurityEvent, 0, len(existing))
	for _, se := range existing {
		if se.timestamp.After(cutoff) {
			result = append(result, se.event)
		}
	}

	return result, nil
}

// GetCount returns the number of active events stored under key within window.
func (c *CorrelationStore) GetCount(ctx context.Context, key string, window time.Duration) (int, error) {
	events, err := c.GetEvents(ctx, key, window)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

// Clear removes all stored state for key.
func (c *CorrelationStore) Clear(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
	return nil
}
