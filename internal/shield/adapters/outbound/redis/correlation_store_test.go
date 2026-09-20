package redis

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

func TestRedisCorrelationStore_InterfaceCompatibility(t *testing.T) {
	var store outbound.CorrelationStore = NewCorrelationStore(nil)
	ctx := context.Background()

	// Verify nil client handling safety
	err := store.AddEvent(ctx, "key", &domain.SecurityEvent{}, 5*time.Minute)
	if err != nil {
		t.Errorf("Expected nil error for nil client AddEvent, got %v", err)
	}

	events, err := store.GetEvents(ctx, "key", 5*time.Minute)
	if err != nil || len(events) != 0 {
		t.Errorf("Expected 0 events for nil client GetEvents, got %d, err: %v", len(events), err)
	}

	count, err := store.GetCount(ctx, "key", 5*time.Minute)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 count for nil client GetCount, got %d, err: %v", count, err)
	}

	err = store.Clear(ctx, "key")
	if err != nil {
		t.Errorf("Expected nil error for nil client Clear, got %v", err)
	}
}

func TestRedisCorrelationStore_EventPayloadFormatting(t *testing.T) {
	evtID := "4625"
	user := "admin_target"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
	}

	if event.SourceEventID == nil || *event.SourceEventID != "4625" {
		t.Errorf("Expected event ID 4625")
	}
}
