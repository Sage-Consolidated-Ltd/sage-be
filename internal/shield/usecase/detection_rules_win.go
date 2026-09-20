package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
)

// Factory function to register all initial Windows Detection Rules into an IncidentEngine.
func RegisterDefaultWindowsRules(engine inboundEngineRegistrer, store outbound.CorrelationStore) error {
	rules := []DetectionRule{
		// Authentication Detections
		NewWinFailedLoginsRule(5, 5*time.Minute),
		&WinSuspiciousLogonRule{},
		NewWinRepeatedAuthFailuresRule(10, 10*time.Minute),

		// Account Activity Detections
		&WinAccountCreationRule{},
		&WinAccountDeletionRule{},
		&WinAccountModificationRule{},
		&WinGroupMembershipChangeRule{},

		// Privilege / Security Detections
		&WinPrivilegeEscalationRule{},
		&WinAuditPolicyChangeRule{},

		// Process / Service Detections
		&WinSuspiciousProcessCreationRule{},
		&WinSuspiciousServiceActivityRule{},
	}

	for _, r := range rules {
		if err := engine.RegisterRule(r); err != nil {
			return err
		}
	}
	return nil
}

type inboundEngineRegistrer interface {
	RegisterRule(rule any) error
}

// -----------------------------------------------------------------------------
// 1. AUTHENTICATION DETECTIONS
// -----------------------------------------------------------------------------

// WinFailedLoginsRule (Stateful): 5+ Event ID 4625 for same TargetUser within window
type WinFailedLoginsRule struct {
	threshold int
	window    time.Duration
}

func NewWinFailedLoginsRule(threshold int, window time.Duration) *WinFailedLoginsRule {
	if threshold <= 0 {
		threshold = 5
	}
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &WinFailedLoginsRule{threshold: threshold, window: window}
}

func (r *WinFailedLoginsRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:             "win_auth_multiple_failed_logins",
		Name:           "Multiple Failed Windows Logins",
		Description:    "Detects multiple authentication failures (Event 4625) for a target user within a sliding time window",
		Category:       domain.RuleCategoryAuthentication,
		Severity:       types.SeverityHigh,
		Enabled:        true,
		IsStateful:     true,
		WindowDuration: r.window,
		ThresholdCount: r.threshold,
	}
}

func (r *WinFailedLoginsRule) EvaluateStateful(ctx context.Context, event *domain.SecurityEvent, store outbound.CorrelationStore) (*domain.DetectionResult, error) {
	if getEventID(event) != "4625" {
		return &domain.DetectionResult{Matched: false}, nil
	}

	targetUser := getActorUser(event)
	sourceIP := getSourceIP(event)
	key := fmt.Sprintf("%s:%s:%s", r.Metadata().ID, event.OrganizationID.String(), targetUser)

	_ = store.AddEvent(ctx, key, event, r.window)
	events, err := store.GetEvents(ctx, key, r.window)
	if err != nil {
		return nil, err
	}

	if len(events) >= r.threshold {
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityHigh,
			Evidence: domain.Evidence{
				RuleID:            r.Metadata().ID,
				TargetUser:        targetUser,
				SourceIP:          sourceIP,
				AttemptCount:      len(events),
				TimeWindowSeconds: int(r.window.Seconds()),
				Context: map[string]interface{}{
					"correlation_key": key,
					"threshold":       r.threshold,
				},
			},
		}, nil
	}

	return &domain.DetectionResult{Matched: false}, nil
}

// WinSuspiciousLogonRule (Stateless): Event ID 4624 with LogonType 10 (RemoteInteractive/RDP) or privileged account
type WinSuspiciousLogonRule struct{}

func (r *WinSuspiciousLogonRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_auth_suspicious_logon",
		Name:        "Suspicious Successful Windows Logon",
		Description: "Detects successful Remote Interactive (RDP) logons (LogonType 10) or suspicious logon parameters",
		Category:    domain.RuleCategoryAuthentication,
		Severity:    types.SeverityMedium,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinSuspiciousLogonRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) != "4624" {
		return &domain.DetectionResult{Matched: false}, nil
	}

	logonType := getPayloadString(event, "logon_type", "LogonType")
	targetUser := getActorUser(event)

	// LogonType 10 = RemoteInteractive (RDP)
	isRDP := logonType == "10"
	isPrivUser := strings.EqualFold(targetUser, "Administrator") || strings.EqualFold(targetUser, "SYSTEM")

	if isRDP || (isPrivUser && logonType != "2" && logonType != "5") {
		sev := types.SeverityMedium
		if isRDP && isPrivUser {
			sev = types.SeverityHigh
		}

		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     sev,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: targetUser,
				SourceIP:   getSourceIP(event),
				Context: map[string]interface{}{
					"logon_type": logonType,
					"is_rdp":     isRDP,
				},
			},
		}, nil
	}

	return &domain.DetectionResult{Matched: false}, nil
}

// WinRepeatedAuthFailuresRule (Stateful): Rapid failure burst across multiple accounts from same source IP
type WinRepeatedAuthFailuresRule struct {
	threshold int
	window    time.Duration
}

func NewWinRepeatedAuthFailuresRule(threshold int, window time.Duration) *WinRepeatedAuthFailuresRule {
	if threshold <= 0 {
		threshold = 10
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &WinRepeatedAuthFailuresRule{threshold: threshold, window: window}
}

func (r *WinRepeatedAuthFailuresRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:             "win_auth_repeated_failures_source_ip",
		Name:           "Repeated Authentication Failures from Source IP",
		Description:    "Detects password spraying or brute force bursts coming from a single source IP address",
		Category:       domain.RuleCategoryAuthentication,
		Severity:       types.SeverityHigh,
		Enabled:        true,
		IsStateful:     true,
		WindowDuration: r.window,
		ThresholdCount: r.threshold,
	}
}

func (r *WinRepeatedAuthFailuresRule) EvaluateStateful(ctx context.Context, event *domain.SecurityEvent, store outbound.CorrelationStore) (*domain.DetectionResult, error) {
	if getEventID(event) != "4625" {
		return &domain.DetectionResult{Matched: false}, nil
	}

	sourceIP := getSourceIP(event)
	if sourceIP == "" || sourceIP == "-" || sourceIP == "127.0.0.1" {
		return &domain.DetectionResult{Matched: false}, nil
	}

	key := fmt.Sprintf("%s:%s:%s", r.Metadata().ID, event.OrganizationID.String(), sourceIP)
	_ = store.AddEvent(ctx, key, event, r.window)

	events, err := store.GetEvents(ctx, key, r.window)
	if err != nil {
		return nil, err
	}

	if len(events) >= r.threshold {
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityHigh,
			Evidence: domain.Evidence{
				RuleID:            r.Metadata().ID,
				SourceIP:          sourceIP,
				TargetUser:        getActorUser(event),
				AttemptCount:      len(events),
				TimeWindowSeconds: int(r.window.Seconds()),
			},
		}, nil
	}

	return &domain.DetectionResult{Matched: false}, nil
}

// -----------------------------------------------------------------------------
// 2. ACCOUNT ACTIVITY DETECTIONS
// -----------------------------------------------------------------------------

// WinAccountCreationRule (Stateless): Event ID 4720 (User account created)
type WinAccountCreationRule struct{}

func (r *WinAccountCreationRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_account_creation",
		Name:        "Windows User Account Created",
		Description: "Detects creation of a new user account (Event 4720)",
		Category:    domain.RuleCategoryAccount,
		Severity:    types.SeverityMedium,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinAccountCreationRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "4720" {
		targetUser := getPayloadString(event, "target_user_name", "TargetUserName")
		if targetUser == "" {
			targetUser = getActorUser(event)
		}
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityMedium,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: targetUser,
				Context: map[string]interface{}{
					"event_id":   "4720",
					"created_by": getActorUser(event),
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// WinAccountDeletionRule (Stateless): Event ID 4726 (User account deleted)
type WinAccountDeletionRule struct{}

func (r *WinAccountDeletionRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_account_deletion",
		Name:        "Windows User Account Deleted",
		Description: "Detects deletion of a user account (Event 4726)",
		Category:    domain.RuleCategoryAccount,
		Severity:    types.SeverityMedium,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinAccountDeletionRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "4726" {
		targetUser := getPayloadString(event, "target_user_name", "TargetUserName")
		if targetUser == "" {
			targetUser = getActorUser(event)
		}
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityMedium,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: targetUser,
				Context: map[string]interface{}{
					"event_id":   "4726",
					"deleted_by": getActorUser(event),
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// WinAccountModificationRule (Stateless): Event ID 4738 (User account modified)
type WinAccountModificationRule struct{}

func (r *WinAccountModificationRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_account_modification",
		Name:        "Windows User Account Modified",
		Description: "Detects modification of user account parameters (Event 4738)",
		Category:    domain.RuleCategoryAccount,
		Severity:    types.SeverityLow,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinAccountModificationRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "4738" {
		targetUser := getPayloadString(event, "target_user_name", "TargetUserName")
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityLow,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: targetUser,
				Context: map[string]interface{}{
					"event_id":    "4738",
					"modified_by": getActorUser(event),
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// WinGroupMembershipChangeRule (Stateless): Event IDs 4728, 4732, 4756 (Member added to security group)
type WinGroupMembershipChangeRule struct{}

func (r *WinGroupMembershipChangeRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_account_group_membership_change",
		Name:        "Privileged Group Membership Change",
		Description: "Detects addition of users to security groups (Events 4728, 4732, 4756)",
		Category:    domain.RuleCategoryAccount,
		Severity:    types.SeverityHigh,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinGroupMembershipChangeRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	evtID := getEventID(event)
	if evtID == "4728" || evtID == "4732" || evtID == "4756" {
		groupName := getPayloadString(event, "target_group_name", "TargetGroupName")
		memberName := getPayloadString(event, "member_name", "MemberName")

		sev := types.SeverityMedium
		if strings.Contains(strings.ToLower(groupName), "admin") {
			sev = types.SeverityHigh
		}

		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     sev,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: memberName,
				Context: map[string]interface{}{
					"event_id":   evtID,
					"group_name": groupName,
					"added_by":   getActorUser(event),
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// -----------------------------------------------------------------------------
// 3. PRIVILEGE & SECURITY POLICY DETECTIONS
// -----------------------------------------------------------------------------

// WinPrivilegeEscalationRule (Stateless): Event ID 4672 (Special privileges assigned to new logon)
type WinPrivilegeEscalationRule struct{}

func (r *WinPrivilegeEscalationRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_privilege_escalation_assigned",
		Name:        "Special Privileges Assigned to Logon",
		Description: "Detects assignment of sensitive administrator privileges (Event 4672)",
		Category:    domain.RuleCategoryPrivilege,
		Severity:    types.SeverityMedium,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinPrivilegeEscalationRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "4672" {
		user := getActorUser(event)

		// Exclude standard system automated accounts if needed
		if user == "SYSTEM" || user == "LOCAL SERVICE" || user == "NETWORK SERVICE" {
			return &domain.DetectionResult{Matched: false}, nil
		}

		privList := getPayloadString(event, "privilege_list", "PrivilegeList")

		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityMedium,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: user,
				Context: map[string]interface{}{
					"event_id":       "4672",
					"privilege_list": privList,
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// WinAuditPolicyChangeRule (Stateless): Event ID 4719 (System audit policy changed)
type WinAuditPolicyChangeRule struct{}

func (r *WinAuditPolicyChangeRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_security_audit_policy_change",
		Name:        "Windows Audit Policy Change",
		Description: "Detects modification or disabling of system audit logging policy (Event 4719)",
		Category:    domain.RuleCategoryPrivilege,
		Severity:    types.SeverityHigh,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinAuditPolicyChangeRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "4719" {
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityHigh,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: getActorUser(event),
				Context: map[string]interface{}{
					"event_id":    "4719",
					"modified_by": getActorUser(event),
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// -----------------------------------------------------------------------------
// 4. PROCESS & SERVICE ACTIVITY DETECTIONS
// -----------------------------------------------------------------------------

// WinSuspiciousProcessCreationRule (Stateless): Event ID 4688 with suspicious process name / CLI
type WinSuspiciousProcessCreationRule struct{}

func (r *WinSuspiciousProcessCreationRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_process_suspicious_creation",
		Name:        "Suspicious Process Creation",
		Description: "Detects execution of suspicious command-line utilities (powershell -enc, whoami, certutil, mimikatz, vssadmin)",
		Category:    domain.RuleCategoryProcessService,
		Severity:    types.SeverityHigh,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinSuspiciousProcessCreationRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) != "4688" {
		return &domain.DetectionResult{Matched: false}, nil
	}

	procName := strings.ToLower(getPayloadString(event, "process_name", "NewProcessName"))
	cli := strings.ToLower(getPayloadString(event, "command_line", "CommandLine"))

	suspiciousKeywords := []string{
		"mimikatz",
		"certutil",
		"-enc",
		"vssadmin delete shadows",
		"whoami /all",
		"lsass.dmp",
		"net localgroup administrators",
		"cmd.exe /c powershell",
	}

	matchedKeyword := ""
	for _, kw := range suspiciousKeywords {
		if strings.Contains(procName, kw) || strings.Contains(cli, kw) {
			matchedKeyword = kw
			break
		}
	}

	if matchedKeyword != "" {
		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     types.SeverityHigh,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: getActorUser(event),
				Context: map[string]interface{}{
					"event_id":        "4688",
					"process_name":    procName,
					"command_line":    cli,
					"matched_keyword": matchedKeyword,
				},
			},
		}, nil
	}

	return &domain.DetectionResult{Matched: false}, nil
}

// WinSuspiciousServiceActivityRule (Stateless): Event ID 7045 (New service installed)
type WinSuspiciousServiceActivityRule struct{}

func (r *WinSuspiciousServiceActivityRule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "win_service_suspicious_installation",
		Name:        "New Windows Service Installed",
		Description: "Detects installation of a new system service (Event 7045)",
		Category:    domain.RuleCategoryProcessService,
		Severity:    types.SeverityMedium,
		Enabled:     true,
		IsStateful:  false,
	}
}

func (r *WinSuspiciousServiceActivityRule) Evaluate(event *domain.SecurityEvent) (*domain.DetectionResult, error) {
	if getEventID(event) == "7045" {
		serviceName := getPayloadString(event, "service_name", "ServiceName")
		imagePath := getPayloadString(event, "image_path", "ImagePath")

		sev := types.SeverityMedium
		if strings.Contains(strings.ToLower(imagePath), "temp") || strings.Contains(strings.ToLower(imagePath), "appdata") {
			sev = types.SeverityHigh
		}

		return &domain.DetectionResult{
			RuleMetadata: r.Metadata(),
			Matched:      true,
			Severity:     sev,
			Evidence: domain.Evidence{
				RuleID:     r.Metadata().ID,
				TargetUser: getActorUser(event),
				Context: map[string]interface{}{
					"event_id":     "7045",
					"service_name": serviceName,
					"image_path":   imagePath,
				},
			},
		}, nil
	}
	return &domain.DetectionResult{Matched: false}, nil
}

// -----------------------------------------------------------------------------
// HELPER UTILITIES FOR PAYLOAD EXTRACTION
// -----------------------------------------------------------------------------

func getEventID(event *domain.SecurityEvent) string {
	if event.SourceEventID != nil && *event.SourceEventID != "" {
		return *event.SourceEventID
	}
	if event.EventType != "" {
		return event.EventType
	}
	if event.NormalizedPayload != nil {
		if id, ok := event.NormalizedPayload["event_id"].(string); ok {
			return id
		}
	}
	return ""
}

func getActorUser(event *domain.SecurityEvent) string {
	if event.ActorUsername != nil && *event.ActorUsername != "" {
		return *event.ActorUsername
	}
	if event.NormalizedPayload != nil {
		if u, ok := event.NormalizedPayload["user"].(string); ok && u != "" {
			return u
		}
		if u, ok := event.NormalizedPayload["target_user_name"].(string); ok && u != "" {
			return u
		}
	}
	return "UNKNOWN"
}

func getSourceIP(event *domain.SecurityEvent) string {
	if event.IPAddress != nil && *event.IPAddress != "" {
		return *event.IPAddress
	}
	if event.NormalizedPayload != nil {
		if ip, ok := event.NormalizedPayload["ip_address"].(string); ok && ip != "" {
			return ip
		}
	}
	return ""
}

func getPayloadString(event *domain.SecurityEvent, normKey, rawKey string) string {
	if event.NormalizedPayload != nil {
		if val, ok := event.NormalizedPayload[normKey].(string); ok && val != "" {
			return val
		}
	}
	if event.RawPayload != nil {
		if val, ok := event.RawPayload[rawKey].(string); ok && val != "" {
			return val
		}
	}
	return ""
}
