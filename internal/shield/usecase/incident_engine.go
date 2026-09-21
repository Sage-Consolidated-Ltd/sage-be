package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// IncidentEngine implements ports.inbound.IncidentEngine for security event detection.
type IncidentEngine struct {
	mu                sync.RWMutex
	rules             []DetectionRule
	correlationStore  outbound.CorrelationStore
	signatureDetector *ThreatSignaturesDetector
	correlationEngine *CorrelationEngine
	alertRepo         outbound.AlertRepository
}

func (e *IncidentEngine) SetAlertRepository(repo outbound.AlertRepository) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alertRepo = repo
}

// NewIncidentEngine initializes an IncidentEngine instance with a correlation state store and detectors.
func NewIncidentEngine(store outbound.CorrelationStore) inbound.IncidentEngine {
	corrStore := NewCorrelationStore()
	return &IncidentEngine{
		rules:             make([]DetectionRule, 0),
		correlationStore:  store,
		signatureDetector: NewThreatSignaturesDetector(),
		correlationEngine: NewCorrelationEngine(corrStore),
	}
}

// NewIncidentEngineWithRedis initializes an IncidentEngine instance backed by Redis for distributed alert sliding windows.
func NewIncidentEngineWithRedis(store outbound.CorrelationStore, redisClient redis.Cmdable) inbound.IncidentEngine {
	corrStore := NewCorrelationStoreWithRedis(redisClient)
	return &IncidentEngine{
		rules:             make([]DetectionRule, 0),
		correlationStore:  store,
		signatureDetector: NewThreatSignaturesDetector(),
		correlationEngine: NewCorrelationEngine(corrStore),
	}
}

// RegisterRule registers a DetectionRule (stateless or stateful) with the engine.
func (e *IncidentEngine) RegisterRule(rule any) error {
	if rule == nil {
		return fmt.Errorf("cannot register nil detection rule")
	}

	detRule, ok := rule.(DetectionRule)
	if !ok {
		return fmt.Errorf("provided rule does not implement DetectionRule interface")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	meta := detRule.Metadata()
	for _, existing := range e.rules {
		if existing.Metadata().ID == meta.ID {
			return fmt.Errorf("rule with ID '%s' is already registered", meta.ID)
		}
	}

	e.rules = append(e.rules, detRule)
	return nil
}

// GetRegisteredRules returns metadata of all currently registered detection rules (INC-001 through INC-010 + custom).
func (e *IncidentEngine) GetRegisteredRules() []domain.RuleMetadata {
	e.mu.RLock()
	defer e.mu.RUnlock()

	formalRules := []domain.RuleMetadata{
		{
			ID:             "INC-001",
			Name:           "Brute Force Account Compromise",
			Description:    "Failed login burst >= 15 in 60s followed by successful auth within 10m on same IP and account.",
			Category:       domain.RuleCategoryInitialAccess,
			Severity:       types.SeverityHigh,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 10 * time.Minute,
			ThresholdCount: 15,
		},
		{
			ID:             "INC-002",
			Name:           "Lateral Movement Following Compromise",
			Description:    "Initial access logon on Host A followed by lateral logon to Host B within 30 minutes.",
			Category:       domain.RuleCategoryLateralMovement,
			Severity:       types.SeverityHigh,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 30 * time.Minute,
		},
		{
			ID:             "INC-003",
			Name:           "Unauthorized Privileged Account Creation",
			Description:    "User account created (4720) and added to Administrators (4732) within 15 minutes.",
			Category:       domain.RuleCategoryPersistence,
			Severity:       types.SeverityHigh,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 15 * time.Minute,
		},
		{
			ID:             "INC-004",
			Name:           "Reconnaissance Burst",
			Description:    "10+ discovery and recon tools executed within 5 minutes on same host. Seeds 4h watchlist.",
			Category:       domain.RuleCategoryDiscovery,
			Severity:       types.SeverityMedium,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 5 * time.Minute,
			ThresholdCount: 10,
		},
		{
			ID:             "INC-005",
			Name:           "Defense Impairment & Pre-Ransomware Prep",
			Description:    "Security defenses disabled and shadow copies destroyed on same host within 20 minutes.",
			Category:       domain.RuleCategoryDefenseEvasion,
			Severity:       types.SeverityCritical,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 20 * time.Minute,
		},
		{
			ID:             "INC-006",
			Name:           "Credential Theft (Mimikatz Activity)",
			Description:    "Multiple Mimikatz dumping or remediation failure signals on same host within 15 minutes.",
			Category:       domain.RuleCategoryCredentialAccess,
			Severity:       types.SeverityCritical,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 15 * time.Minute,
			ThresholdCount: 2,
		},
		{
			ID:             "INC-007",
			Name:           "Persistence Mechanism Established",
			Description:    "2 or more distinct persistence mechanisms created on same host within 60 minutes.",
			Category:       domain.RuleCategoryPersistence,
			Severity:       types.SeverityHigh,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 60 * time.Minute,
			ThresholdCount: 2,
		},
		{
			ID:             "INC-008",
			Name:           "C2 Channel Established",
			Description:    "Outbound C2 connection correlated with LOLBin execution on same host within 10 minutes.",
			Category:       domain.RuleCategoryInitialAccess,
			Severity:       types.SeverityCritical,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 10 * time.Minute,
		},
		{
			ID:             "INC-009",
			Name:           "Data Exfiltration Pipeline",
			Description:    "Sequential file collection, archive compression, and external upload within 30 minutes.",
			Category:       domain.RuleCategoryExfiltration,
			Severity:       types.SeverityCritical,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 30 * time.Minute,
		},
		{
			ID:             "INC-010",
			Name:           "Confirmed Breach (Meta-Incident)",
			Description:    "3 or more distinct incidents on connected entity graph within 24 hours merged into P1 breach.",
			Category:       domain.RuleCategoryMetaBreach,
			Severity:       types.SeverityCritical,
			Enabled:        true,
			IsStateful:     true,
			WindowDuration: 24 * time.Hour,
			ThresholdCount: 3,
		},
	}

	for _, r := range e.rules {
		formalRules = append(formalRules, r.Metadata())
	}

	return formalRules
}

// EvaluateEvent evaluates a security event against threat signatures and correlation rules.
func (e *IncidentEngine) EvaluateEvent(ctx context.Context, event *domain.SecurityEvent) ([]*domain.Incident, error) {
	if event == nil {
		return []*domain.Incident{}, nil
	}

	// 1. Dual-Inspection Threat Signature Detection
	alerts := e.signatureDetector.DetectAlerts(event)

	// Persist detected alerts to PostgreSQL
	if e.alertRepo != nil && len(alerts) > 0 {
		_ = e.alertRepo.BulkSaveAlerts(ctx, alerts)
	}

	// 2. Correlation Engine (INC-001 through INC-009)
	incidents := e.correlationEngine.EvaluateRules(event.OrganizationID, alerts)

	// 3. Evaluate legacy/custom rules if registered
	legacyIncidents, err := e.evaluateLegacyRules(ctx, event)
	if err == nil && len(legacyIncidents) > 0 {
		incidents = append(incidents, legacyIncidents...)
	}

	// 4. INC-010 Meta-Rule: Breach Merge
	incidents = e.evaluateMetaIncidents(event.OrganizationID, incidents)

	// 5. Dynamic scoring and priority
	for _, inc := range incidents {
		e.applyDynamicScoring(inc)
	}

	// 6. Deduplication
	incidents = e.deduplicateIncidents(incidents)

	return incidents, nil
}

// EvaluateBatch evaluates a slice of events against threat signatures and correlation rules.
func (e *IncidentEngine) EvaluateBatch(ctx context.Context, events []*domain.SecurityEvent) ([]*domain.Incident, error) {
	if len(events) == 0 {
		return nil, nil
	}

	orgID := events[0].OrganizationID
	var validEvents []*domain.SecurityEvent
	for _, ev := range events {
		if ev != nil {
			validEvents = append(validEvents, ev)
		}
	}
	if len(validEvents) == 0 {
		return nil, nil
	}

	orgID := validEvents[0].OrganizationID

	// 1. Collect alerts across all events in batch
	var allAlerts []*domain.Alert
	for _, event := range events {
	for _, event := range validEvents {
		alerts := e.signatureDetector.DetectAlerts(event)
		allAlerts = append(allAlerts, alerts...)
	}

	// Persist detected alerts to PostgreSQL
	if e.alertRepo != nil && len(allAlerts) > 0 {
		_ = e.alertRepo.BulkSaveAlerts(ctx, allAlerts)
	}

	// 2. Correlation rules evaluation
	incidents := e.correlationEngine.EvaluateRules(orgID, allAlerts)

	// 3. Legacy rules evaluation
	for _, event := range events {
	for _, event := range validEvents {
		legacy, _ := e.evaluateLegacyRules(ctx, event)
		if len(legacy) > 0 {
			incidents = append(incidents, legacy...)
		}
	}

	// 4. INC-010 Meta-Rule Breach Merge
	incidents = e.evaluateMetaIncidents(orgID, incidents)

	// 5. Dynamic scoring and priority
	for _, inc := range incidents {
		e.applyDynamicScoring(inc)
		if inc != nil {
			e.applyDynamicScoring(inc)
		}
	}

	// 6. Deduplication
	incidents = e.deduplicateIncidents(incidents)

	return incidents, nil
}

// evaluateLegacyRules evaluates any legacy rules registered via RegisterRule.
func (e *IncidentEngine) evaluateLegacyRules(ctx context.Context, event *domain.SecurityEvent) ([]*domain.Incident, error) {
	e.mu.RLock()
	rules := make([]DetectionRule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	var incidents []*domain.Incident
	for _, rule := range rules {
		meta := rule.Metadata()
		if !meta.Enabled {
			continue
		}

		var detResult *domain.DetectionResult
		var err error

		if meta.IsStateful {
			statefulRule, ok := rule.(StatefulRule)
			if !ok {
				continue
			}
			detResult, err = statefulRule.EvaluateStateful(ctx, event, e.correlationStore)
		} else {
			statelessRule, ok := rule.(StatelessRule)
			if !ok {
				continue
			}
			detResult, err = statelessRule.Evaluate(event)
		}

		if err != nil {
			continue
		}

		if detResult != nil && detResult.Matched {
			inc := e.buildLegacyIncident(event, meta, detResult)
			incidents = append(incidents, inc)
		}
	}

	return incidents, nil
}

// evaluateMetaIncidents merges incidents when >= 3 distinct incident stages occur on an entity within 24h.
func (e *IncidentEngine) evaluateMetaIncidents(orgID uuid.UUID, incidents []*domain.Incident) []*domain.Incident {
	if len(incidents) < 3 {
		return incidents
	}

	entityIncidents := make(map[string][]*domain.Incident)
	for _, inc := range incidents {
		if inc == nil {
			continue
		}
		if inc.EntityKey != "" {
			entityIncidents[inc.EntityKey] = append(entityIncidents[inc.EntityKey], inc)
		}
	}

	var finalIncidents []*domain.Incident
	mergedEntities := make(map[string]bool)

	for entityKey, group := range entityIncidents {
		distinctRules := make(map[string]bool)
		for _, inc := range group {
			if inc == nil {
				continue
			}
			distinctRules[inc.RuleID] = true
		}

		if len(distinctRules) >= 3 {
			mergedEntities[entityKey] = true

			var allSummaries []domain.ContributingEventSummary
			var allMITRE []string
			var allThreats []string
			var latestOccurred time.Time

			for _, inc := range group {
				if inc == nil {
					continue
				}
				allSummaries = append(allSummaries, inc.Evidence.ContributingEvents...)
				allMITRE = append(allMITRE, inc.Evidence.MITRETechniques...)
				allThreats = append(allThreats, inc.Evidence.ContributingThreats...)
				if inc.OccurredAt.After(latestOccurred) {
					latestOccurred = inc.OccurredAt
				}
			}

			metaEvidence := domain.Evidence{
				RuleID:              "INC-010",
				RuleName:            "Confirmed Breach (Meta-Incident)",
				Severity:            types.SeverityCritical,
				Score:               100,
				Priority:            "P1",
				EntityKey:           entityKey,
				HostName:            group[0].Evidence.HostName,
				TargetUser:          group[0].Evidence.TargetUser,
				SourceIP:            group[0].Evidence.SourceIP,
				OccurredAt:          latestOccurred,
				AttemptCount:        len(group),
				TimeWindowSeconds:   86400,
				MITRETechniques:     dedupStrings(allMITRE),
				ContributingThreats: dedupStrings(allThreats),
				ContributingEvents:  allSummaries,
			}

			metaInc := &domain.Incident{
				ID:             uuid.New(),
				OrganizationID: orgID,
				RuleID:         "INC-010",
				RuleName:       "Confirmed Breach (Meta-Incident)",
				Category:       domain.RuleCategoryMetaBreach,
				Severity:       types.SeverityCritical,
				Score:          100,
				Priority:       "P1",
				EntityKey:      entityKey,
				Status:         domain.IncidentStatusNew,
				Title:          fmt.Sprintf("Meta-Incident: Confirmed Multi-Stage Intrusion Breach on %s", entityKey),
				Summary:        fmt.Sprintf("Entity %s exhibited %d distinct incident stages within 24 hours. Breach confirmed.", entityKey, len(distinctRules)),
				Evidence:       metaEvidence,
				OccurredAt:     latestOccurred,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			finalIncidents = append(finalIncidents, metaInc)
		}
	}

	for _, inc := range incidents {
		if inc == nil {
			continue
		}
		if !mergedEntities[inc.EntityKey] {
			finalIncidents = append(finalIncidents, inc)
		}
	}

	return finalIncidents
}

// applyDynamicScoring calculates dynamic score (0-100) and priority (P1/P2/P3).
func (e *IncidentEngine) applyDynamicScoring(inc *domain.Incident) {
	if inc == nil {
		return
	}
	if inc.Score == 0 {
		switch inc.Severity {
		case types.SeverityCritical:
			inc.Score = 85
		case types.SeverityHigh:
			inc.Score = 65
		case types.SeverityMedium:
			inc.Score = 40
		default:
			inc.Score = 20
		}
	}

	host := inc.Evidence.HostName
	if host != "" {
		// +10 if on 4h host watchlist
		if e.correlationEngine.Store().IsHostOnWatchlist(host) {
			inc.Score += 10
		}
		// +15 if anti-forensics crash or time change context detected
		if e.correlationEngine.Store().HasAntiForensicsContext(host, 4*time.Hour) {
			inc.Score += 15
		}
	}

	if inc.Score > 100 {
		inc.Score = 100
	}

	if inc.Score >= 90 || inc.Severity == types.SeverityCritical {
		inc.Priority = "P1"
	} else if inc.Score >= 65 {
		inc.Priority = "P2"
	} else {
		inc.Priority = "P3"
	}

	inc.Evidence.Score = inc.Score
	inc.Evidence.Priority = inc.Priority
}

// deduplicateIncidents collapses duplicate rule firings on the same entity.
func (e *IncidentEngine) deduplicateIncidents(incidents []*domain.Incident) []*domain.Incident {
	type dedupKey struct {
		ruleID    string
		entityKey string
	}

	dedupMap := make(map[dedupKey]*domain.Incident)
	for _, inc := range incidents {
		if inc == nil {
			continue
		}
		k := dedupKey{ruleID: inc.RuleID, entityKey: inc.EntityKey}
		if existing, exists := dedupMap[k]; exists {
			existing.Evidence.ContributingEvents = append(existing.Evidence.ContributingEvents, inc.Evidence.ContributingEvents...)
			if inc.Score > existing.Score {
				existing.Score = inc.Score
				existing.Priority = inc.Priority
				existing.Evidence.Score = inc.Score
				existing.Evidence.Priority = inc.Priority
			}
			if inc.OccurredAt.After(existing.OccurredAt) {
				existing.OccurredAt = inc.OccurredAt
			}
			existing.UpdatedAt = time.Now()
		} else {
			dedupMap[k] = inc
		}
	}

	result := make([]*domain.Incident, 0, len(dedupMap))
	for _, inc := range dedupMap {
		result = append(result, inc)
	}
	return result
}

func (e *IncidentEngine) buildLegacyIncident(
	event *domain.SecurityEvent,
	meta domain.RuleMetadata,
	detResult *domain.DetectionResult,
) *domain.Incident {
	now := time.Now()
	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}

	title := fmt.Sprintf("[%s] %s Detected", meta.Severity, meta.Name)
	summary := meta.Description

	evidence := detResult.Evidence
	evidence.RuleID = meta.ID
	evidence.RuleName = meta.Name
	evidence.Severity = detResult.Severity
	evidence.Description = meta.Description

	if evidence.TriggerEventID == "" {
		if event.SourceEventID != nil {
			evidence.TriggerEventID = *event.SourceEventID
		} else {
			evidence.TriggerEventID = event.ID.String()
		}
	}

	if evidence.TargetUser == "" && event.ActorUsername != nil {
		evidence.TargetUser = *event.ActorUsername
	}
	if evidence.SourceIP == "" && event.IPAddress != nil {
		evidence.SourceIP = *event.IPAddress
	}

	return &domain.Incident{
		ID:             uuid.New(),
		OrganizationID: event.OrganizationID,
		RuleID:         meta.ID,
		RuleName:       meta.Name,
		Category:       meta.Category,
		Severity:       detResult.Severity,
		Status:         domain.IncidentStatusNew,
		Title:          title,
		Summary:        summary,
		Evidence:       evidence,
		OccurredAt:     occurredAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func dedupStrings(list []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range list {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
