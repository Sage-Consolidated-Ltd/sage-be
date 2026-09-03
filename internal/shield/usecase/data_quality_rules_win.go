package usecase

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"

	"github.com/redis/go-redis/v9"
)

var (
	// Windows SID regex standard: S-1-5-...
	sidRegex = regexp.MustCompile(`^S-1-[0-59]-\d{2,}(-\d+)+$`)
)

// isWindowsEvent detects whether a security event originated from a Windows log source.
func isWindowsEvent(event *domain.SecurityEvent) bool {
	if event == nil {
		return false
	}
	src := strings.ToLower(event.Source)
	if src == "" || strings.Contains(src, "win") || strings.Contains(src, "eventlog") || strings.Contains(src, "security") || strings.Contains(src, "system") {
		return true
	}
	if event.NormalizedPayload != nil {
		if _, ok := event.NormalizedPayload["channel"]; ok {
			return true
		}
		if _, ok := event.NormalizedPayload["event_id"]; ok {
			return true
		}
	}
	return false
}

// QualityRule defines the interface for an individual quality check rule.
type QualityRule interface {
	ID() string
	Name() string
	Description() string
	AppliesTo(event *domain.SecurityEvent) bool
	Evaluate(event *domain.SecurityEvent) []domain.QualityIssue
}

// 1. Missing Required Fields Rule
type WinMissingRequiredFieldsRule struct{}

func (r *WinMissingRequiredFieldsRule) ID() string   { return "win_quality_missing_required_fields" }
func (r *WinMissingRequiredFieldsRule) Name() string { return "Missing Required Windows Log Fields" }
func (r *WinMissingRequiredFieldsRule) Description() string {
	return "Verifies essential Windows log fields (Event ID, OccurredAt, Source/Computer, Actor) are present"
}
func (r *WinMissingRequiredFieldsRule) AppliesTo(event *domain.SecurityEvent) bool {
	return isWindowsEvent(event)
}

func (r *WinMissingRequiredFieldsRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.SourceEventID == nil || strings.TrimSpace(*event.SourceEventID) == "" {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_EVENT_ID",
			Message:        "Windows Event ID is missing or blank",
			Severity:       domain.QualitySeverityCritical,
			Category:       "missing_field",
			AffectedFields: []string{"source_event_id"},
		})
	}

	if event.OccurredAt.IsZero() {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_TIMESTAMP",
			Message:        "Event timestamp (occurred_at) is missing or zero",
			Severity:       domain.QualitySeverityCritical,
			Category:       "missing_field",
			AffectedFields: []string{"occurred_at"},
		})
	}

	if event.Source == "" {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_SOURCE",
			Message:        "Event log source identifier is empty",
			Severity:       domain.QualitySeverityError,
			Category:       "missing_field",
			AffectedFields: []string{"source"},
		})
	}

	actorMissing := (event.ActorUsername == nil || strings.TrimSpace(*event.ActorUsername) == "") &&
		(event.ActorEmail == nil || strings.TrimSpace(*event.ActorEmail) == "")

	if actorMissing {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_ACTOR_INFO",
			Message:        "Neither username nor user email is populated for the event actor",
			Severity:       domain.QualitySeverityWarning,
			Category:       "missing_field",
			AffectedFields: []string{"actor_username", "actor_email"},
		})
	}

	return issues
}

// 2. Invalid Fields & Types Rule
type WinInvalidFieldsRule struct{}

func (r *WinInvalidFieldsRule) ID() string   { return "win_quality_invalid_fields" }
func (r *WinInvalidFieldsRule) Name() string { return "Invalid Field Values and Types" }
func (r *WinInvalidFieldsRule) Description() string {
	return "Detects invalid category names, event types, or severe type mismatches"
}
func (r *WinInvalidFieldsRule) AppliesTo(event *domain.SecurityEvent) bool {
	return true
}

func (r *WinInvalidFieldsRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.EventCategory != "" && !event.EventCategory.IsValid() {
		issues = append(issues, domain.QualityIssue{
			Code:           "INVALID_EVENT_CATEGORY",
			Message:        fmt.Sprintf("Unknown or invalid event category '%s'", event.EventCategory),
			Severity:       domain.QualitySeverityError,
			Category:       "invalid_field",
			AffectedFields: []string{"event_category"},
		})
	}

	if event.Severity != nil && !event.Severity.IsValid() {
		issues = append(issues, domain.QualityIssue{
			Code:           "INVALID_SEVERITY_LEVEL",
			Message:        fmt.Sprintf("Invalid severity level '%s'", *event.Severity),
			Severity:       domain.QualitySeverityError,
			Category:       "invalid_field",
			AffectedFields: []string{"severity"},
		})
	}

	return issues
}

// 3. Malformed Values Rule (IPs, SIDs)
type WinMalformedValuesRule struct{}

func (r *WinMalformedValuesRule) ID() string   { return "win_quality_malformed_values" }
func (r *WinMalformedValuesRule) Name() string { return "Malformed Value Formats" }
func (r *WinMalformedValuesRule) Description() string {
	return "Checks format validity for IP addresses, Windows Security Identifiers (SIDs), and GUIDs"
}
func (r *WinMalformedValuesRule) AppliesTo(event *domain.SecurityEvent) bool {
	return true
}

func (r *WinMalformedValuesRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.IPAddress != nil && strings.TrimSpace(*event.IPAddress) != "" && *event.IPAddress != "-" {
		ipStr := strings.TrimSpace(*event.IPAddress)
		if net.ParseIP(ipStr) == nil {
			issues = append(issues, domain.QualityIssue{
				Code:           "MALFORMED_IP_ADDRESS",
				Message:        fmt.Sprintf("IP address '%s' is not a valid IPv4 or IPv6 string", ipStr),
				Severity:       domain.QualitySeverityError,
				Category:       "malformed_value",
				AffectedFields: []string{"ip_address"},
			})
		}
	}

	// Check SIDs in normalized or raw payload if present
	if event.NormalizedPayload != nil {
		if sidVal, exists := event.NormalizedPayload["user_sid"]; exists {
			if sidStr, ok := sidVal.(string); ok && sidStr != "" && sidStr != "-" {
				if !sidRegex.MatchString(sidStr) && !strings.HasPrefix(sidStr, "S-1-") {
					issues = append(issues, domain.QualityIssue{
						Code:           "MALFORMED_WINDOWS_SID",
						Message:        fmt.Sprintf("Windows SID '%s' has an invalid structure", sidStr),
						Severity:       domain.QualitySeverityWarning,
						Category:       "malformed_value",
						AffectedFields: []string{"normalized_payload.user_sid"},
					})
				}
			}
		}
	}

	return issues
}

// 4. Invalid Timestamp Rule
type WinInvalidTimestampRule struct{}

func (r *WinInvalidTimestampRule) ID() string   { return "win_quality_invalid_timestamp" }
func (r *WinInvalidTimestampRule) Name() string { return "Invalid or Out-of-Bounds Timestamp" }
func (r *WinInvalidTimestampRule) Description() string {
	return "Detects timestamps in the future (>5m), zero timestamps, or stale events older than 365 days"
}
func (r *WinInvalidTimestampRule) AppliesTo(event *domain.SecurityEvent) bool {
	return true
}

func (r *WinInvalidTimestampRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue
	now := time.Now()

	if event.OccurredAt.IsZero() {
		return issues // Handled by missing fields rule
	}

	if event.OccurredAt.After(now.Add(5 * time.Minute)) {
		issues = append(issues, domain.QualityIssue{
			Code:           "FUTURE_TIMESTAMP",
			Message:        fmt.Sprintf("Event timestamp %s is in the future (> 5m threshold)", event.OccurredAt.Format(time.RFC3339)),
			Severity:       domain.QualitySeverityCritical,
			Category:       "invalid_timestamp",
			AffectedFields: []string{"occurred_at"},
		})
	}

	if event.OccurredAt.Before(now.AddDate(-1, 0, 0)) {
		issues = append(issues, domain.QualityIssue{
			Code:           "STALE_TIMESTAMP",
			Message:        fmt.Sprintf("Event timestamp %s is older than 365 days", event.OccurredAt.Format(time.RFC3339)),
			Severity:       domain.QualitySeverityWarning,
			Category:       "invalid_timestamp",
			AffectedFields: []string{"occurred_at"},
		})
	}

	return issues
}

// 5. Missing Windows Metadata Rule (Channel, Provider, Task)
type WinMissingMetadataRule struct{}

func (r *WinMissingMetadataRule) ID() string   { return "win_quality_missing_metadata" }
func (r *WinMissingMetadataRule) Name() string { return "Missing Windows Event Metadata" }
func (r *WinMissingMetadataRule) Description() string {
	return "Checks for Windows Event Log metadata like Channel, ProviderName, and Task Category"
}
func (r *WinMissingMetadataRule) AppliesTo(event *domain.SecurityEvent) bool {
	return isWindowsEvent(event)
}

func (r *WinMissingMetadataRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	hasChannel := false
	hasProvider := false

	if event.NormalizedPayload != nil {
		if ch, ok := event.NormalizedPayload["channel"].(string); ok && ch != "" {
			hasChannel = true
		}
		if pr, ok := event.NormalizedPayload["provider"].(string); ok && pr != "" {
			hasProvider = true
		}
	}
	if event.RawPayload != nil {
		if ch, ok := event.RawPayload["Channel"].(string); ok && ch != "" {
			hasChannel = true
		}
		if pr, ok := event.RawPayload["ProviderName"].(string); ok && pr != "" {
			hasProvider = true
		}
	}

	if !hasChannel {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_WIN_CHANNEL",
			Message:        "Windows Log Channel metadata (Security, System, Application) is missing",
			Severity:       domain.QualitySeverityWarning,
			Category:       "missing_metadata",
			AffectedFields: []string{"channel"},
		})
	}

	if !hasProvider {
		issues = append(issues, domain.QualityIssue{
			Code:           "MISSING_WIN_PROVIDER",
			Message:        "Windows ProviderName metadata is missing",
			Severity:       domain.QualitySeverityWarning,
			Category:       "missing_metadata",
			AffectedFields: []string{"provider"},
		})
	}

	return issues
}

// 6. Parse Status & Error Check Rule
type WinParseStatusRule struct{}

func (r *WinParseStatusRule) ID() string   { return "win_quality_parse_status" }
func (r *WinParseStatusRule) Name() string { return "Parsing & Normalization Quality Check" }
func (r *WinParseStatusRule) Description() string {
	return "Identifies failed parsing status or embedded parse errors in normalized payload"
}
func (r *WinParseStatusRule) AppliesTo(event *domain.SecurityEvent) bool {
	return true
}

func (r *WinParseStatusRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.ParseStatus == types.ParseStatusFailed {
		issues = append(issues, domain.QualityIssue{
			Code:           "PARSING_FAILED",
			Message:        "Event failed parser execution and could not be properly normalized",
			Severity:       domain.QualitySeverityCritical,
			Category:       "parse_error",
			AffectedFields: []string{"parse_status"},
		})
	} else if event.ParseStatus == types.ParseStatusPartial {
		issues = append(issues, domain.QualityIssue{
			Code:           "PARSING_PARTIAL",
			Message:        "Event was only partially parsed with warnings or missing tokens",
			Severity:       domain.QualitySeverityWarning,
			Category:       "parse_error",
			AffectedFields: []string{"parse_status"},
		})
	}

	if len(event.ParseErrors) > 0 {
		issues = append(issues, domain.QualityIssue{
			Code:           "PARSE_ERRORS_RECORDED",
			Message:        fmt.Sprintf("Event contains %d parse error records", len(event.ParseErrors)),
			Severity:       domain.QualitySeverityError,
			Category:       "parse_error",
			AffectedFields: []string{"parse_errors"},
		})
	}

	return issues
}

// 7. Duplicate Event Detection Rule
type WinDuplicateEventRule struct {
	mu           sync.Mutex
	seenHashes   map[string]time.Time
	windowPeriod time.Duration
	redisClient  *redis.Client
}

func NewWinDuplicateEventRule(windowPeriod time.Duration) *WinDuplicateEventRule {
	return NewWinDuplicateEventRuleWithRedis(windowPeriod, nil)
}

func NewWinDuplicateEventRuleWithRedis(windowPeriod time.Duration, redisClient *redis.Client) *WinDuplicateEventRule {
	if windowPeriod <= 0 {
		windowPeriod = 10 * time.Minute
	}
	return &WinDuplicateEventRule{
		seenHashes:   make(map[string]time.Time),
		windowPeriod: windowPeriod,
		redisClient:  redisClient,
	}
}

func (r *WinDuplicateEventRule) ID() string   { return "win_quality_duplicate_event" }
func (r *WinDuplicateEventRule) Name() string { return "Duplicate Windows Event Check" }
func (r *WinDuplicateEventRule) Description() string {
	return "Detects duplicate events arriving within a sliding deduplication time window"
}
func (r *WinDuplicateEventRule) AppliesTo(event *domain.SecurityEvent) bool {
	return true
}

func (r *WinDuplicateEventRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.SourceEventID == nil || event.OccurredAt.IsZero() {
		return issues
	}

	hashKey := fmt.Sprintf("%s|%s|%s|%d",
		event.OrganizationID.String(),
		*event.SourceEventID,
		event.Source,
		event.OccurredAt.UnixNano(),
	)

	if r.redisClient != nil {
		key := fmt.Sprintf("quality:seen:%s", hashKey)
		set, err := r.redisClient.SetNX(context.Background(), key, "1", r.windowPeriod).Result()
		if err == nil && !set {
			issues = append(issues, domain.QualityIssue{
				Code:           "DUPLICATE_EVENT_DETECTED",
				Message:        fmt.Sprintf("Event with duplicate key '%s' was already processed within %s", hashKey, r.windowPeriod),
				Severity:       domain.QualitySeverityError,
				Category:       "duplicate",
				AffectedFields: []string{"source_event_id", "occurred_at"},
			})
		}
		return issues
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Prune stale cache entries periodically
	for k, t := range r.seenHashes {
		if now.Sub(t) > r.windowPeriod {
			delete(r.seenHashes, k)
		}
	}

	if seenAt, exists := r.seenHashes[hashKey]; exists && now.Sub(seenAt) <= r.windowPeriod {
		issues = append(issues, domain.QualityIssue{
			Code:           "DUPLICATE_EVENT_DETECTED",
			Message:        fmt.Sprintf("Event with duplicate key '%s' was already processed within %s", hashKey, r.windowPeriod),
			Severity:       domain.QualitySeverityError,
			Category:       "duplicate",
			AffectedFields: []string{"source_event_id", "occurred_at"},
		})
	} else {
		r.seenHashes[hashKey] = now
	}

	return issues
}

// 8. Partially Populated Event Rule
type WinPartialPopulationRule struct{}

func (r *WinPartialPopulationRule) ID() string   { return "win_quality_partial_population" }
func (r *WinPartialPopulationRule) Name() string { return "Partially Populated Event Field Check" }
func (r *WinPartialPopulationRule) Description() string {
	return "Verifies payload populated field ratio against standard Windows event expectations"
}
func (r *WinPartialPopulationRule) AppliesTo(event *domain.SecurityEvent) bool {
	return isWindowsEvent(event)
}

func (r *WinPartialPopulationRule) Evaluate(event *domain.SecurityEvent) []domain.QualityIssue {
	var issues []domain.QualityIssue

	if event.NormalizedPayload == nil || len(event.NormalizedPayload) == 0 {
		issues = append(issues, domain.QualityIssue{
			Code:           "EMPTY_NORMALIZED_PAYLOAD",
			Message:        "Normalized payload is empty or nil",
			Severity:       domain.QualitySeverityError,
			Category:       "partial_population",
			AffectedFields: []string{"normalized_payload"},
		})
		return issues
	}

	expectedStandardKeys := []string{"event_id", "channel", "computer", "user", "ip_address", "process_name"}
	foundCount := 0
	for _, key := range expectedStandardKeys {
		if val, ok := event.NormalizedPayload[key]; ok && val != nil && val != "" {
			foundCount++
		}
	}

	// If fewer than 2 standard fields populated out of expected
	if foundCount < 2 {
		issues = append(issues, domain.QualityIssue{
			Code:           "PARTIAL_POPULATION_LOW_COVERAGE",
			Message:        fmt.Sprintf("Event normalized payload contains only %d/%d standard Windows fields", foundCount, len(expectedStandardKeys)),
			Severity:       domain.QualitySeverityWarning,
			Category:       "partial_population",
			AffectedFields: []string{"normalized_payload"},
		})
	}

	return issues
}
