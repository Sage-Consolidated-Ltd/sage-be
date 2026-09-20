package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/adapters/outbound/memory"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

func setupTestEngine() (*IncidentEngine, context.Context) {
	store := memory.NewCorrelationStore()
	engine := NewIncidentEngine(store).(*IncidentEngine)
	_ = RegisterDefaultWindowsRules(engine, store)
	return engine, context.Background()
}

func TestWindowsRules_RegisterDefaultSuite(t *testing.T) {
	engine, _ := setupTestEngine()
	rules := engine.GetRegisteredRules()
	if len(rules) < 10 {
		t.Fatalf("Expected at least 10 default Windows rules, got %d", len(rules))
	}
}

func TestWindowsRules_FailedLoginsStateful(t *testing.T) {
	engine, ctx := setupTestEngine()
	orgID := uuid.New()
	user := "target_admin"
	evtID := "4625"

	// 4 attempts -> no incident
	for i := 0; i < 4; i++ {
		event := &domain.SecurityEvent{
			ID:             uuid.New(),
			OrganizationID: orgID,
			SourceEventID:  &evtID,
			ActorUsername:  &user,
			OccurredAt:     time.Now(),
		}
		incidents, _ := engine.EvaluateEvent(ctx, event)
		if len(incidents) != 0 {
			t.Fatalf("Attempt %d: expected 0 incidents before threshold", i+1)
		}
	}

	// 5th attempt -> High severity incident
	fifthEvent := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		SourceEventID:  &evtID,
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
	}

	incidents, err := engine.EvaluateEvent(ctx, fifthEvent)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(incidents) == 0 {
		t.Fatalf("Expected incident on 5th failed login attempt")
	}

	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_auth_multiple_failed_logins" {
			found = true
			if inc.Severity != types.SeverityHigh {
				t.Errorf("Expected High severity, got %s", inc.Severity)
			}
		}
	}
	if !found {
		t.Errorf("win_auth_multiple_failed_logins incident not found")
	}
}

func TestWindowsRules_SuspiciousLogonRDP(t *testing.T) {
	engine, ctx := setupTestEngine()
	evtID := "4624"
	user := "Administrator"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"logon_type": "10",
		},
	}

	incidents, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_auth_suspicious_logon" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected win_auth_suspicious_logon incident for RDP admin logon")
	}
}

func TestWindowsRules_AccountLifecycle(t *testing.T) {
	engine, ctx := setupTestEngine()

	// 1. Account Created (4720)
	evtID20 := "4720"
	createEvt := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID20,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"target_user_name": "temp_contractor",
		},
	}
	incidents1, _ := engine.EvaluateEvent(ctx, createEvt)
	if len(incidents1) == 0 || incidents1[0].RuleID != "win_account_creation" {
		t.Errorf("Expected win_account_creation incident")
	}

	// 2. Account Deleted (4726)
	evtID26 := "4726"
	deleteEvt := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID26,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"target_user_name": "temp_contractor",
		},
	}
	incidents2, _ := engine.EvaluateEvent(ctx, deleteEvt)
	if len(incidents2) == 0 || incidents2[0].RuleID != "win_account_deletion" {
		t.Errorf("Expected win_account_deletion incident")
	}
}

func TestWindowsRules_GroupMembershipChange(t *testing.T) {
	engine, ctx := setupTestEngine()
	evtID := "4728"
	adminUser := "attacker_user"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"target_group_name": "Domain Admins",
			"member_name":       adminUser,
		},
	}

	incidents, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_account_group_membership_change" {
			found = true
			if inc.Severity != types.SeverityHigh {
				t.Errorf("Expected High severity for Domain Admins group change, got %s", inc.Severity)
			}
		}
	}
	if !found {
		t.Errorf("Expected win_account_group_membership_change incident")
	}
}

func TestWindowsRules_AuditPolicyChange(t *testing.T) {
	engine, ctx := setupTestEngine()
	evtID := "4719"
	user := "rogue_admin"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
	}

	incidents, _ := engine.EvaluateEvent(ctx, event)
	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_security_audit_policy_change" {
			found = true
			if inc.Severity != types.SeverityHigh {
				t.Errorf("Expected High severity for Audit Policy Change")
			}
		}
	}
	if !found {
		t.Errorf("Expected win_security_audit_policy_change incident")
	}
}

func TestWindowsRules_SuspiciousProcessCreation(t *testing.T) {
	engine, ctx := setupTestEngine()
	evtID := "4688"
	user := "compromised_user"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		ActorUsername:  &user,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"process_name": "C:\\Windows\\System32\\powershell.exe",
			"command_line": "powershell.exe -enc SQBFAFgAIAAoAE4AZQB3AC0ATwBiAGoAZQBjAHQ...==",
		},
	}

	incidents, _ := engine.EvaluateEvent(ctx, event)
	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_process_suspicious_creation" {
			found = true
			if inc.Severity != types.SeverityHigh {
				t.Errorf("Expected High severity for encoded powershell command line")
			}
		}
	}
	if !found {
		t.Errorf("Expected win_process_suspicious_creation incident")
	}
}

func TestWindowsRules_ServiceInstallation(t *testing.T) {
	engine, ctx := setupTestEngine()
	evtID := "7045"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &evtID,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"service_name": "PersistenceSvc",
			"image_path":   "C:\\Users\\Public\\AppData\\Local\\Temp\\malware.exe",
		},
	}

	incidents, _ := engine.EvaluateEvent(ctx, event)
	found := false
	for _, inc := range incidents {
		if inc.RuleID == "win_service_suspicious_installation" {
			found = true
			if inc.Severity != types.SeverityHigh {
				t.Errorf("Expected High severity for Temp directory service installation")
			}
		}
	}
	if !found {
		t.Errorf("Expected win_service_suspicious_installation incident")
	}
}
