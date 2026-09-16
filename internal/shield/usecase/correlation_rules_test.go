package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 1. Dual-Inspection Unit Test
func TestDualInspection_ParsedAndUnparsed(t *testing.T) {
	detector := NewThreatSignaturesDetector()
	orgID := uuid.New()
	user := "administrator"
	host := "WORKSTATION-1"
	ip := "192.168.1.100"

	// A. Normalized event (ParseStatusSuccess)
	parsedEvt := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Security",
		EventType:      "4625",
		ActorUsername:  &user,
		IPAddress:      &ip,
		ParseStatus:    types.ParseStatusSuccess,
		NormalizedPayload: map[string]interface{}{
			"Computer":  host,
			"LogonType": 3,
		},
		OccurredAt: time.Now(),
	}

	alertsA := detector.DetectAlerts(parsedEvt)
	require.Len(t, alertsA, 1)
	assert.Equal(t, "Brute_Force_Failed_Login", alertsA[0].ThreatLabel)
	assert.Equal(t, "T1110.001", alertsA[0].MITRE)
	assert.Equal(t, "4625", alertsA[0].EventID)
	assert.Equal(t, user, alertsA[0].EntityAccount)
	assert.Equal(t, host, alertsA[0].EntityHost)
	assert.Equal(t, ip, alertsA[0].EntityIP)

	// B. Unparsed polled event (ParseStatusPending with raw JSON payload)
	unparsedEvt := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Security",
		EventType:      "windows:4625",
		ParseStatus:    types.ParseStatusPending,
		RawPayload: map[string]interface{}{
			"EventID":        "4625",
			"TargetUserName": user,
			"ComputerName":   host,
			"IpAddress":      ip,
			"LogonType":      "3",
		},
		OccurredAt: time.Now(),
	}

	alertsB := detector.DetectAlerts(unparsedEvt)
	require.Len(t, alertsB, 1)
	assert.Equal(t, "Brute_Force_Failed_Login", alertsB[0].ThreatLabel)
	assert.Equal(t, user, alertsB[0].EntityAccount)
	assert.Equal(t, host, alertsB[0].EntityHost)
	assert.Equal(t, ip, alertsB[0].EntityIP)
}

// 2. INC-001: Brute Force Account Compromise
func TestINC001_BruteForceAccountCompromise(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	account := "finance_user"
	ip := "203.0.113.50"
	host := "FIN-DESKTOP-01"

	var alerts []*domain.Alert
	// 15 failed logins within 30 seconds
	for i := 0; i < 15; i++ {
		alerts = append(alerts, &domain.Alert{
			ID:            uuid.New(),
			ThreatLabel:   "Brute_Force_Failed_Login",
			MITRE:         "T1110.001",
			LogSource:     "Security",
			EventID:       "4625",
			EntityHost:    host,
			EntityAccount: account,
			EntityIP:      ip,
			DetectedAt:    now.Add(time.Duration(i*2) * time.Second),
		})
	}

	// Successful auth 2 minutes later
	alerts = append(alerts, &domain.Alert{
		ID:            uuid.New(),
		ThreatLabel:   "Brute_Force_Successful_Auth",
		MITRE:         "T1078.003",
		LogSource:     "Security",
		EventID:       "4624",
		EntityHost:    host,
		EntityAccount: account,
		EntityIP:      ip,
		DetectedAt:    now.Add(2 * time.Minute),
	})

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-001", incidents[0].RuleID)
	assert.Equal(t, types.SeverityHigh, incidents[0].Severity)
	assert.Equal(t, "account:finance_user", incidents[0].EntityKey)
	assert.Equal(t, 65, incidents[0].Score)
	assert.Equal(t, "P2", incidents[0].Priority)

	// Test escalation with Account Lockout (4740)
	alertsWithLockout := append(alerts, &domain.Alert{
		ID:            uuid.New(),
		ThreatLabel:   "Account_Lockout",
		MITRE:         "T1110.001",
		LogSource:     "Security",
		EventID:       "4740",
		EntityHost:    host,
		EntityAccount: account,
		EntityIP:      ip,
		DetectedAt:    now.Add(45 * time.Second),
	})

	store2 := NewCorrelationStore()
	engine2 := NewCorrelationEngine(store2)
	incidents2 := engine2.EvaluateRules(orgID, alertsWithLockout)
	require.Len(t, incidents2, 1)
	assert.Equal(t, types.SeverityCritical, incidents2[0].Severity)
	assert.Equal(t, 85, incidents2[0].Score)
	assert.Equal(t, "P1", incidents2[0].Priority)
}

// 3. INC-002: Lateral Movement Following Compromise
func TestINC002_LateralMovement(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	account := "compromised_admin"

	alerts := []*domain.Alert{
		{
			ID:            uuid.New(),
			ThreatLabel:   "Brute_Force_Successful_Auth",
			LogSource:     "Security",
			EventID:       "4624",
			EntityHost:    "WORKSTATION-1",
			EntityAccount: account,
			EntityIP:      "192.168.1.10",
			DetectedAt:    now,
		},
		{
			ID:            uuid.New(),
			ThreatLabel:   "Lateral_Movement_RDP",
			LogSource:     "Security",
			EventID:       "4624",
			EntityHost:    "DC-SERVER-01",
			EntityAccount: account,
			EntityIP:      "192.168.1.10",
			DetectedAt:    now.Add(15 * time.Minute),
		},
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-002", incidents[0].RuleID)
	assert.Equal(t, types.SeverityHigh, incidents[0].Severity)
	assert.Contains(t, incidents[0].Title, "DC-SERVER-01")
}

// 4. INC-003: Unauthorized Privileged Account Creation
func TestINC003_PrivilegedAccountCreation(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	account := "backdoor_admin"

	alerts := []*domain.Alert{
		{
			ID:            uuid.New(),
			ThreatLabel:   "New_User_Account_Created",
			EventID:       "4720",
			EntityHost:    "DC-01",
			EntityAccount: account,
			DetectedAt:    now,
		},
		{
			ID:            uuid.New(),
			ThreatLabel:   "User_Added_To_Administrators",
			EventID:       "4732",
			EntityHost:    "DC-01",
			EntityAccount: account,
			DetectedAt:    now.Add(5 * time.Minute),
		},
		{
			ID:            uuid.New(),
			ThreatLabel:   "User_Account_Deleted",
			EventID:       "4726",
			EntityHost:    "DC-01",
			EntityAccount: account,
			DetectedAt:    now.Add(1 * time.Hour),
		},
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-003", incidents[0].RuleID)
	// Base 65 + 15 (deleted cleanup boost) = 80
	assert.Equal(t, 80, incidents[0].Score)
	assert.Contains(t, incidents[0].Evidence.ContributingThreats, "User_Account_Deleted")
}

// 5. INC-004: Reconnaissance Burst & Watchlist
func TestINC004_ReconBurst_Watchlist(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	host := "TARGET-HOST-01"

	var alerts []*domain.Alert
	for i := 0; i < 12; i++ {
		alerts = append(alerts, &domain.Alert{
			ID:          uuid.New(),
			ThreatLabel: "Recon_Process_Chain",
			EventID:     "4688",
			EntityHost:  host,
			DetectedAt:  now.Add(time.Duration(i*10) * time.Second),
		})
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-004", incidents[0].RuleID)
	assert.Equal(t, types.SeverityMedium, incidents[0].Severity)
	assert.True(t, store.IsHostOnWatchlist(host), "Host must be placed on 4-hour watchlist")
}

// 6. INC-005: Defense Impairment & Pre-Ransomware Prep
func TestINC005_PreRansomware(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	host := "FILE-SERVER-01"

	alerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Defense_Disabled_Defender",
			EventID:     "13",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Shadow_Copy_Destruction",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(10 * time.Minute),
		},
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-005", incidents[0].RuleID)
	assert.Equal(t, types.SeverityCritical, incidents[0].Severity)
	assert.Equal(t, 85, incidents[0].Score)
	assert.Equal(t, "P1", incidents[0].Priority)
}

// 7. INC-006: Credential Theft (Mimikatz Activity)
func TestINC006_CredentialTheft(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	host := "DEV-MACHINE-05"

	alerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Mimikatz_Dropped",
			EventID:     "11",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "PowerShell_Mimikatz_Command",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(3 * time.Minute),
		},
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-006", incidents[0].RuleID)
	assert.Equal(t, types.SeverityCritical, incidents[0].Severity)
}

// 8. INC-007: Persistence Diversity
func TestINC007_PersistenceDiversity(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	host := "WORKSTATION-99"

	// Same technique repeated twice should NOT trigger INC-007
	sameTechAlerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Persistence_RunKey",
			EventID:     "13",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Persistence_RunKey",
			EventID:     "13",
			EntityHost:  host,
			DetectedAt:  now.Add(5 * time.Minute),
		},
	}
	incidentsSame := engine.EvaluateRules(orgID, sameTechAlerts)
	assert.Empty(t, incidentsSame, "Same persistence technique repeated should not trigger INC-007")

	// Two DISTINCT techniques trigger INC-007
	distinctAlerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Persistence_RunKey",
			EventID:     "13",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Persistence_Malicious_Service",
			EventID:     "7045",
			EntityHost:  host,
			DetectedAt:  now.Add(10 * time.Minute),
		},
	}
	incidentsDistinct := engine.EvaluateRules(orgID, distinctAlerts)
	require.Len(t, incidentsDistinct, 1)
	assert.Equal(t, "INC-007", incidentsDistinct[0].RuleID)
}

// 9. INC-008: C2 Channel Established
func TestINC008_C2Channel(t *testing.T) {
	orgID := uuid.New()
	store := NewCorrelationStore()
	engine := NewCorrelationEngine(store)
	now := time.Now()
	host := "LAPTOP-CORP"

	alerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Outbound_C2_Connection",
			EventID:     "3",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "LOLBin_Execution",
			EventID:     "1",
			EntityHost:  host,
			DetectedAt:  now.Add(4 * time.Minute),
		},
	}

	incidents := engine.EvaluateRules(orgID, alerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-008", incidents[0].RuleID)
	assert.Equal(t, types.SeverityCritical, incidents[0].Severity)
}

// 10. INC-009: Data Exfiltration Pipeline (Strict Ordering)
func TestINC009_DataExfiltration_Ordered(t *testing.T) {
	orgID := uuid.New()
	host := "RESEARCH-SERVER"
	now := time.Now()

	// Out of order: Upload happened BEFORE Collection -> should NOT trigger
	outOfOrderAlerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Exfiltration_Upload",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Collection_Files",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(5 * time.Minute),
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Archive_Compression",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(10 * time.Minute),
		},
	}
	store1 := NewCorrelationStore()
	engine1 := NewCorrelationEngine(store1)
	assert.Empty(t, engine1.EvaluateRules(orgID, outOfOrderAlerts))

	// In order: Collect -> Archive -> Upload within 30m -> Triggers INC-009
	inOrderAlerts := []*domain.Alert{
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Collection_Files",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now,
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Archive_Compression",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(5 * time.Minute),
		},
		{
			ID:          uuid.New(),
			ThreatLabel: "Data_Exfiltration_Upload",
			EventID:     "4104",
			EntityHost:  host,
			DetectedAt:  now.Add(12 * time.Minute),
		},
	}
	store2 := NewCorrelationStore()
	engine2 := NewCorrelationEngine(store2)
	incidents := engine2.EvaluateRules(orgID, inOrderAlerts)
	require.Len(t, incidents, 1)
	assert.Equal(t, "INC-009", incidents[0].RuleID)
	assert.Equal(t, types.SeverityCritical, incidents[0].Severity)
}

// 11. INC-010: Confirmed Breach (Meta-Incident)
func TestINC010_ConfirmedBreachMetaMerge(t *testing.T) {
	incidentEngine := NewIncidentEngine(nil)
	orgID := uuid.New()
	host := "COMPROMISED-HOST-01"
	now := time.Now()

	// Feed events to trigger 3 distinct rules on the same host:
	// 1) INC-004 (10 recon events)
	// 2) INC-005 (Defense disabled + shadow copy destruction)
	// 3) INC-006 (Mimikatz dropped + PS mimikatz command)
	var events []*domain.SecurityEvent

	// INC-004
	for i := 0; i < 11; i++ {
		events = append(events, &domain.SecurityEvent{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Source:         "Security",
			EventType:      "4688",
			OccurredAt:     now.Add(time.Duration(i*5) * time.Second),
			RawPayload: map[string]interface{}{
				"CommandLine": "whoami /all",
				"Computer":    host,
			},
		})
	}

	// INC-005
	events = append(events, &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Sysmon",
		EventType:      "13",
		OccurredAt:     now.Add(1 * time.Minute),
		RawPayload: map[string]interface{}{
			"EventID":      "13",
			"TargetObject": `services\windefend`,
			"Details":      "start=4",
			"Computer":     host,
		},
	})
	events = append(events, &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Application",
		EventType:      "4104",
		OccurredAt:     now.Add(2 * time.Minute),
		RawPayload: map[string]interface{}{
			"EventID":         "4104",
			"ScriptBlockText": "vssadmin delete shadows /all",
			"Computer":        host,
		},
	})

	// INC-006
	events = append(events, &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Sysmon",
		EventType:      "11",
		OccurredAt:     now.Add(3 * time.Minute),
		RawPayload: map[string]interface{}{
			"EventID":        "11",
			"TargetFilename": `C:\Windows\Temp\mimikatz.exe`,
			"Computer":       host,
		},
	})
	events = append(events, &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Source:         "Application",
		EventType:      "4104",
		OccurredAt:     now.Add(4 * time.Minute),
		RawPayload: map[string]interface{}{
			"EventID":         "4104",
			"ScriptBlockText": "Invoke-Mimikatz -Command sekurlsa::logonpasswords",
			"Computer":        host,
		},
	})

	incidents, err := incidentEngine.EvaluateBatch(context.Background(), events)
	require.NoError(t, err)

	// Verify INC-010 meta-incident was produced
	var metaInc *domain.Incident
	for _, inc := range incidents {
		if inc.RuleID == "INC-010" {
			metaInc = inc
			break
		}
	}
	require.NotNil(t, metaInc, "Expected INC-010 Confirmed Breach meta-incident to be produced")
	assert.Equal(t, 100, metaInc.Score)
	assert.Equal(t, "P1", metaInc.Priority)
	assert.Equal(t, types.SeverityCritical, metaInc.Severity)
	assert.Contains(t, metaInc.Title, "Confirmed Multi-Stage Intrusion Breach")
}

