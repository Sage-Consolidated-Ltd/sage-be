package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/adapters/outbound/memory"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

// Mock Stateless Rule
type mockStatelessRule struct {
	meta domain.RuleMetadata
}

func (r *mockStatelessRule) Metadata() domain.RuleMetadata { return r.meta }
func (r *mockStatelessRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if event.EventType == "4720" { // Account created
		user := "johndoe"
		if event.ActorUsername != nil {
			user = *event.ActorUsername
		}
		return &domain.DetectionResult{
			RuleMetadata: r.meta,
			Matched:      true,
			Severity:     r.meta.Severity,
			Evidence: domain.Evidence{
				RuleID:     r.meta.ID,
				TargetUser: user,
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// Mock Stateful Rule
type mockStatefulRule struct {
	meta      domain.RuleMetadata
	threshold int
}

func (r *mockStatefulRule) Metadata() domain.RuleMetadata { return r.meta }
func (r *mockStatefulRule) EvaluateStateful(ctx context.Context, event *domain.SecurityEvent, store outbound.CorrelationStore) (*domain.DetectionResult, error) {
	if event.EventType != "4625" { // Failed login
		return &domain.DetectionResult{Matched: false}, nil
	}

	user := "unknown"
	if event.ActorUsername != nil {
		user = *event.ActorUsername
	}

	key := r.meta.ID + ":" + user
	_ = store.AddEvent(ctx, key, event, 5*time.Minute)

	events, _ := store.GetEvents(ctx, key, 5*time.Minute)
	if len(events) >= r.threshold {
		return &domain.DetectionResult{
			RuleMetadata: r.meta,
			Matched:      true,
			Severity:     r.meta.Severity,
			Evidence: domain.Evidence{
				RuleID:            r.meta.ID,
				TargetUser:        user,
				AttemptCount:      len(events),
				TimeWindowSeconds: 300,
			},
		}, nil
	}

	return &domain.DetectionResult{Matched: false}, nil
}

func TestIncidentEngine_StatelessRule(t *testing.T) {
	store := memory.NewCorrelationStore()
	engine := NewIncidentEngine(store).(*IncidentEngine)
	ctx := context.Background()

	rule := &mockStatelessRule{
		meta: domain.RuleMetadata{
			ID:          "rule_acc_create",
			Name:        "Windows Account Created",
			Description: "Detects creation of new Windows user accounts",
			Category:    domain.RuleCategoryAccount,
			Severity:    types.SeverityMedium,
			Enabled:     true,
			IsStateful:  false,
		},
	}

	if err := engine.RegisterRule(rule); err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	evtID := "4720"
	user := "new_admin"
	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		EventType:      "4720",
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
	}

	incidents, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected evaluation error: %v", err)
	}

	if len(incidents) != 1 {
		t.Fatalf("Expected 1 incident, got %d", len(incidents))
	}

	inc := incidents[0]
	if inc.RuleID != "rule_acc_create" {
		t.Errorf("Expected RuleID 'rule_acc_create', got '%s'", inc.RuleID)
	}
	if inc.Severity != types.SeverityMedium {
		t.Errorf("Expected Medium severity, got '%s'", inc.Severity)
	}
	if inc.Evidence.TargetUser != "new_admin" {
		t.Errorf("Expected Evidence TargetUser 'new_admin', got '%s'", inc.Evidence.TargetUser)
	}
}

func TestIncidentEngine_StatefulRuleCorrelation(t *testing.T) {
	store := memory.NewCorrelationStore()
	engine := NewIncidentEngine(store).(*IncidentEngine)
	ctx := context.Background()

	rule := &mockStatefulRule{
		meta: domain.RuleMetadata{
			ID:             "rule_failed_logins",
			Name:           "Multiple Failed Logins",
			Description:    "Detects 3+ failed logins within 5 minutes",
			Category:       domain.RuleCategoryAuthentication,
			Severity:       types.SeverityHigh,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 5 * time.Minute,
			ThresholdCount: 3,
		},
		threshold: 3,
	}

	_ = engine.RegisterRule(rule)

	evtID := "4625"
	targetUser := "victim_user"
	now := time.Now()

	// Send 2 failed attempts -> No incident yet
	for i := 0; i < 2; i++ {
		event := &domain.SecurityEvent{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			SourceEventID:  &evtID,
			EventType:      "4625",
			ActorUsername:  &targetUser,
			OccurredAt:     now,
		}
		incidents, _ := engine.EvaluateEvent(ctx, event)
		if len(incidents) != 0 {
			t.Fatalf("Attempt %d: Expected 0 incidents before threshold, got %d", i+1, len(incidents))
		}
	}

	// 3rd attempt -> Should trigger high severity incident
	thirdEvent := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		EventType:      "4625",
		ActorUsername:  &targetUser,
		OccurredAt:     now,
	}

	incidents, err := engine.EvaluateEvent(ctx, thirdEvent)
	if err != nil {
		t.Fatalf("Unexpected error on 3rd attempt: %v", err)
	}

	if len(incidents) != 1 {
		t.Fatalf("Expected 1 incident on 3rd attempt, got %d", len(incidents))
	}

	inc := incidents[0]
	if inc.Severity != types.SeverityHigh {
		t.Errorf("Expected High severity incident, got %s", inc.Severity)
	}
	if inc.Evidence.AttemptCount != 3 {
		t.Errorf("Expected evidence AttemptCount = 3, got %d", inc.Evidence.AttemptCount)
	}
}
