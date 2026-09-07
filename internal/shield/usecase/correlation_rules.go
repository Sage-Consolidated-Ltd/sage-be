package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CorrelationStore holds active alert sliding windows, host watchlists, and anti-forensics tracking.
// Supports Redis-backed distributed state with in-memory fallback for local dev and tests.
type CorrelationStore struct {
	mu            sync.RWMutex
	alerts        []*domain.Alert
	hostWatchlist map[string]time.Time // host -> watchlist expiration time (4h from INC-004)
	antiForensics map[string]time.Time // host -> crash/time change occurred
	redisClient   redis.Cmdable
	prefix        string
}

func NewCorrelationStore() *CorrelationStore {
	return &CorrelationStore{
		alerts:        make([]*domain.Alert, 0),
		hostWatchlist: make(map[string]time.Time),
		antiForensics: make(map[string]time.Time),
		prefix:        "corr:",
	}
}

func NewCorrelationStoreWithRedis(client redis.Cmdable) *CorrelationStore {
	store := NewCorrelationStore()
	store.redisClient = client
	return store
}

func (s *CorrelationStore) AddAlert(alert *domain.Alert) {
	if alert == nil {
		return
	}

	s.mu.Lock()
	s.alerts = append(s.alerts, alert)
	if alert.ThreatLabel == "System_Crash_AntiForensics" || alert.ThreatLabel == "System_Time_Changed" {
		if alert.EntityHost != "" {
			s.antiForensics[alert.EntityHost] = alert.DetectedAt
		}
	}
	s.mu.Unlock()

	if s.redisClient != nil {
		ctx := context.Background()
		orgKey := "global"
		if alert.OrganizationID != uuid.Nil {
			orgKey = alert.OrganizationID.String()
		}
		redisKey := fmt.Sprintf("%salerts:%s", s.prefix, orgKey)
		score := float64(alert.DetectedAt.UnixNano())
		b, err := json.Marshal(alert)
		if err == nil {
			pipe := s.redisClient.TxPipeline()
			pipe.ZAdd(ctx, redisKey, redis.Z{Score: score, Member: string(b)})
			cutoff := float64(time.Now().Add(-24 * time.Hour).UnixNano())
			pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("(%f", cutoff))
			pipe.Expire(ctx, redisKey, 24*time.Hour)
			_, _ = pipe.Exec(ctx)
		}

		if alert.ThreatLabel == "System_Crash_AntiForensics" || alert.ThreatLabel == "System_Time_Changed" {
			if alert.EntityHost != "" {
				antiKey := fmt.Sprintf("%santiforensics:%s:%s", s.prefix, orgKey, alert.EntityHost)
				_ = s.redisClient.Set(ctx, antiKey, "1", 4*time.Hour).Err()
			}
		}
	}
}

func (s *CorrelationStore) AddToHostWatchlist(host string, duration time.Duration) {
	s.mu.Lock()
	s.hostWatchlist[host] = time.Now().Add(duration)
	s.mu.Unlock()

	if s.redisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf("%swatchlist:%s", s.prefix, host)
		_ = s.redisClient.Set(ctx, key, "1", duration).Err()
	}
}

func (s *CorrelationStore) IsHostOnWatchlist(host string) bool {
	if s.redisClient != nil {
		ctx := context.Background()
		key := fmt.Sprintf("%swatchlist:%s", s.prefix, host)
		val, err := s.redisClient.Exists(ctx, key).Result()
		if err == nil && val > 0 {
			return true
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if exp, ok := s.hostWatchlist[host]; ok {
		return time.Now().Before(exp)
	}
	return false
}

func (s *CorrelationStore) HasAntiForensicsContext(host string, window time.Duration) bool {
	if s.redisClient != nil {
		ctx := context.Background()
		keys, err := s.redisClient.Keys(ctx, fmt.Sprintf("%santiforensics:*:%s", s.prefix, host)).Result()
		if err == nil && len(keys) > 0 {
			return true
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if ts, ok := s.antiForensics[host]; ok {
		return time.Since(ts) <= window
	}
	return false
}

func (s *CorrelationStore) GetAlertsInWindow(since time.Time) []*domain.Alert {
	if s.redisClient != nil {
		ctx := context.Background()
		minScore := strconv.FormatInt(since.UnixNano(), 10)
		keys, err := s.redisClient.Keys(ctx, fmt.Sprintf("%salerts:*", s.prefix)).Result()
		if err == nil && len(keys) > 0 {
			var result []*domain.Alert
			for _, k := range keys {
				rawAlerts, zErr := s.redisClient.ZRangeByScore(ctx, k, &redis.ZRangeBy{
					Min: minScore,
					Max: "+inf",
				}).Result()
				if zErr == nil {
					for _, raw := range rawAlerts {
						var a domain.Alert
						if json.Unmarshal([]byte(raw), &a) == nil {
							result = append(result, &a)
						}
					}
				}
			}
			if len(result) > 0 {
				return result
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.Alert
	for _, a := range s.alerts {
		if a.DetectedAt.After(since) || a.DetectedAt.Equal(since) {
			result = append(result, a)
		}
	}
	return result
}

func (s *CorrelationStore) PurgeOlderThan(cutoff time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var active []*domain.Alert
	for _, a := range s.alerts {
		if a.DetectedAt.After(cutoff) {
			active = append(active, a)
		}
	}
	s.alerts = active

	for h, exp := range s.hostWatchlist {
		if time.Now().After(exp) {
			delete(s.hostWatchlist, h)
		}
	}
	for h, ts := range s.antiForensics {
		if time.Since(ts) > 24*time.Hour {
			delete(s.antiForensics, h)
		}
	}
}

// CorrelationEngine orchestrates evaluating INC-001 through INC-009 on incoming alerts.
type CorrelationEngine struct {
	store *CorrelationStore
}

func NewCorrelationEngine(store *CorrelationStore) *CorrelationEngine {
	if store == nil {
		store = NewCorrelationStore()
	}
	return &CorrelationEngine{store: store}
}

func (e *CorrelationEngine) Store() *CorrelationStore {
	return e.store
}

// EvaluateRules evaluates all 9 atomic correlation rules against active alert history and candidate alerts.
func (e *CorrelationEngine) EvaluateRules(orgID uuid.UUID, newAlerts []*domain.Alert) []*domain.Incident {
	for _, a := range newAlerts {
		e.store.AddAlert(a)
	}

	var incidents []*domain.Incident

	// Evaluate INC-001 through INC-009
	if inc := e.evaluateINC001(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC002(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC003(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC004(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC005(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC006(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC007(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC008(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}
	if inc := e.evaluateINC009(orgID); inc != nil {
		incidents = append(incidents, inc...)
	}

	return incidents
}

// INC-001: Brute Force Account Compromise
// Failed login burst >= 15 in 60s -> Successful auth within 10m on same IP + account. Escalate if 4740 present.
func (e *CorrelationEngine) evaluateINC001(orgID uuid.UUID) []*domain.Incident {
	window := 15 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	// Group by Account and IP
	type key struct {
		account string
		ip      string
	}
	failedLogins := make(map[key][]*domain.Alert)
	successLogins := make(map[key][]*domain.Alert)
	lockouts := make(map[string][]*domain.Alert) // by account

	for _, a := range alerts {
		k := key{account: a.EntityAccount, ip: a.EntityIP}
		switch a.ThreatLabel {
		case "Brute_Force_Failed_Login":
			if a.EntityAccount != "" {
				failedLogins[k] = append(failedLogins[k], a)
			}
		case "Brute_Force_Successful_Auth":
			if a.EntityAccount != "" {
				successLogins[k] = append(successLogins[k], a)
			}
		case "Account_Lockout":
			if a.EntityAccount != "" {
				lockouts[a.EntityAccount] = append(lockouts[a.EntityAccount], a)
			}
		}
	}

	var incidents []*domain.Incident
	for k, failures := range failedLogins {
		if len(failures) < 15 {
			continue
		}
		// Sort failures by time
		sort.Slice(failures, func(i, j int) bool {
			return failures[i].DetectedAt.Before(failures[j].DetectedAt)
		})

		// Check if burst of 15 happened within 60s
		hasBurst := false
		var burstEnd time.Time
		for i := 0; i <= len(failures)-15; i++ {
			if failures[i+14].DetectedAt.Sub(failures[i].DetectedAt) <= 60*time.Second {
				hasBurst = true
				burstEnd = failures[i+14].DetectedAt
				break
			}
		}
		if !hasBurst {
			continue
		}

		// Check for subsequent successful login within 10m
		successes := successLogins[k]
		var matchedSuccess *domain.Alert
		for _, s := range successes {
			if s.DetectedAt.After(burstEnd) && s.DetectedAt.Sub(burstEnd) <= 10*time.Minute {
				matchedSuccess = s
				break
			}
		}
		if matchedSuccess == nil {
			continue
		}

		// Determine severity and escalation
		sev := types.SeverityHigh
		score := 65
		hasLockout := len(lockouts[k.account]) > 0
		if hasLockout {
			sev = types.SeverityCritical
			score = 85
		}

		// Watchlist & extra signal boosts
		host := matchedSuccess.EntityHost
		if e.store.IsHostOnWatchlist(host) {
			score += 10
		}
		if len(failures) > 25 {
			score += 10
		}
		if score > 100 {
			score = 100
		}

		priority := "P2"
		if score >= 90 || sev == types.SeverityCritical {
			priority = "P1"
		}

		ev := buildEvidence("INC-001", "Brute Force Account Compromise", sev, score, priority,
			fmt.Sprintf("account:%s", k.account), k.account, k.ip, host, matchedSuccess.DetectedAt,
			len(failures)+1, int(window.Seconds()), []string{"T1110.001", "T1078.003"},
			[]string{"Brute_Force_Failed_Login", "Brute_Force_Successful_Auth"},
			failures, matchedSuccess)

		incidents = append(incidents, &domain.Incident{
			ID:             uuid.New(),
			OrganizationID: orgID,
			RuleID:         "INC-001",
			RuleName:       "Brute Force Account Compromise",
			Category:       domain.RuleCategoryInitialAccess,
			Severity:       sev,
			Score:          score,
			Priority:       priority,
			EntityKey:      fmt.Sprintf("account:%s", k.account),
			Status:         domain.IncidentStatusNew,
			Title:          fmt.Sprintf("Brute Force Compromise against account '%s'", k.account),
			Summary:        fmt.Sprintf("Account %s suffered %d failed logins within 60s from %s followed by successful authentication.", k.account, len(failures), k.ip),
			Evidence:       ev,
			OccurredAt:     matchedSuccess.DetectedAt,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	return incidents
}

// INC-002: Lateral Movement Following Compromise
// Initial access auth on Host A -> lateral logon (LogonType 10, 3+NTLM, 9) on Host B within 30m.
func (e *CorrelationEngine) evaluateINC002(orgID uuid.UUID) []*domain.Incident {
	window := 30 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	authsByAccount := make(map[string][]*domain.Alert)
	lateralByAccount := make(map[string][]*domain.Alert)

	for _, a := range alerts {
		if a.EntityAccount == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Brute_Force_Successful_Auth", "Privileged_Account_Logon":
			authsByAccount[a.EntityAccount] = append(authsByAccount[a.EntityAccount], a)
		case "Lateral_Movement_RDP", "Lateral_Movement_SMB", "Lateral_Movement_NewCredentials":
			lateralByAccount[a.EntityAccount] = append(lateralByAccount[a.EntityAccount], a)
		}
	}

	var incidents []*domain.Incident
	for acc, laterals := range lateralByAccount {
		auths := authsByAccount[acc]
		if len(auths) == 0 {
			continue
		}

		for _, lat := range laterals {
			for _, initial := range auths {
				// Must be different hosts and lateral must occur after initial within 30m
				if initial.EntityHost != "" && lat.EntityHost != "" && initial.EntityHost != lat.EntityHost &&
					lat.DetectedAt.After(initial.DetectedAt) && lat.DetectedAt.Sub(initial.DetectedAt) <= 30*time.Minute {

					score := 65
					if e.store.IsHostOnWatchlist(lat.EntityHost) || e.store.IsHostOnWatchlist(initial.EntityHost) {
						score += 10
					}
					priority := "P2"
					if score >= 90 {
						priority = "P1"
					}

					ev := buildEvidence("INC-002", "Lateral Movement Following Compromise", types.SeverityHigh, score, priority,
						fmt.Sprintf("account:%s", acc), acc, lat.EntityIP, lat.EntityHost, lat.DetectedAt,
						2, int(window.Seconds()), []string{"T1021.001", "T1021.002", "T1078"},
						[]string{initial.ThreatLabel, lat.ThreatLabel},
						[]*domain.Alert{initial, lat})

					incidents = append(incidents, &domain.Incident{
						ID:             uuid.New(),
						OrganizationID: orgID,
						RuleID:         "INC-002",
						RuleName:       "Lateral Movement Following Compromise",
						Category:       domain.RuleCategoryLateralMovement,
						Severity:       types.SeverityHigh,
						Score:          score,
						Priority:       priority,
						EntityKey:      fmt.Sprintf("account:%s", acc),
						Status:         domain.IncidentStatusNew,
						Title:          fmt.Sprintf("Lateral Movement for account '%s' to '%s'", acc, lat.EntityHost),
						Summary:        fmt.Sprintf("Account %s authenticated on %s and moved laterally to %s within %v.", acc, initial.EntityHost, lat.EntityHost, lat.DetectedAt.Sub(initial.DetectedAt)),
						Evidence:       ev,
						OccurredAt:     lat.DetectedAt,
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
					})
					break
				}
			}
		}
	}

	return incidents
}

// INC-003: Unauthorized Privileged Account Creation
// 4720 (User Created) -> 4732 (Added to Administrators) within 15m. Optional 4726 (Deleted) in 4h.
func (e *CorrelationEngine) evaluateINC003(orgID uuid.UUID) []*domain.Incident {
	window := 4 * time.Hour
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	creates := make(map[string]*domain.Alert)
	promotions := make(map[string]*domain.Alert)
	deletions := make(map[string]*domain.Alert)

	for _, a := range alerts {
		if a.EntityAccount == "" {
			continue
		}
		switch a.ThreatLabel {
		case "New_User_Account_Created":
			creates[a.EntityAccount] = a
		case "User_Added_To_Administrators":
			promotions[a.EntityAccount] = a
		case "User_Account_Deleted":
			deletions[a.EntityAccount] = a
		}
	}

	var incidents []*domain.Incident
	for acc, createAlert := range creates {
		promoAlert, ok := promotions[acc]
		if !ok {
			continue
		}
		// Must be within 15m
		if promoAlert.DetectedAt.Sub(createAlert.DetectedAt) <= 15*time.Minute &&
			promoAlert.DetectedAt.Sub(createAlert.DetectedAt) >= 0 {

			score := 65
			contributing := []string{"New_User_Account_Created", "User_Added_To_Administrators"}
			contributingAlerts := []*domain.Alert{createAlert, promoAlert}

			// Context boost if deleted in 4h (anti-forensics backdoor cleanup)
			if delAlert, deleted := deletions[acc]; deleted && delAlert.DetectedAt.After(promoAlert.DetectedAt) {
				score += 15
				contributing = append(contributing, "User_Account_Deleted")
				contributingAlerts = append(contributingAlerts, delAlert)
			}
			if e.store.IsHostOnWatchlist(createAlert.EntityHost) {
				score += 10
			}
			if score > 100 {
				score = 100
			}

			priority := "P2"
			if score >= 90 {
				priority = "P1"
			}

			ev := buildEvidence("INC-003", "Unauthorized Privileged Account Creation", types.SeverityHigh, score, priority,
				fmt.Sprintf("account:%s", acc), acc, createAlert.EntityIP, createAlert.EntityHost, promoAlert.DetectedAt,
				len(contributingAlerts), int(window.Seconds()), []string{"T1136.001", "T1098"},
				contributing, contributingAlerts)

			incidents = append(incidents, &domain.Incident{
				ID:             uuid.New(),
				OrganizationID: orgID,
				RuleID:         "INC-003",
				RuleName:       "Unauthorized Privileged Account Creation",
				Category:       domain.RuleCategoryPersistence,
				Severity:       types.SeverityHigh,
				Score:          score,
				Priority:       priority,
				EntityKey:      fmt.Sprintf("account:%s", acc),
				Status:         domain.IncidentStatusNew,
				Title:          fmt.Sprintf("Backdoor Administrator Account Created: '%s'", acc),
				Summary:        fmt.Sprintf("Account %s was created and escalated to Administrators group within %v.", acc, promoAlert.DetectedAt.Sub(createAlert.DetectedAt)),
				Evidence:       ev,
				OccurredAt:     promoAlert.DetectedAt,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}
	}

	return incidents
}

// INC-004: Reconnaissance Burst
// 10+ recon tools in 5m on same host. Seeds 4-hour host watchlist.
func (e *CorrelationEngine) evaluateINC004(orgID uuid.UUID) []*domain.Incident {
	window := 5 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	reconsByHost := make(map[string][]*domain.Alert)
	for _, a := range alerts {
		if a.ThreatLabel == "Recon_Process_Chain" && a.EntityHost != "" {
			reconsByHost[a.EntityHost] = append(reconsByHost[a.EntityHost], a)
		}
	}

	var incidents []*domain.Incident
	for host, list := range reconsByHost {
		if len(list) >= 10 {
			// Seed 4-hour host watchlist
			e.store.AddToHostWatchlist(host, 4*time.Hour)

			score := 40
			if len(list) > 20 {
				score += 10
			}
			priority := "P3"

			ev := buildEvidence("INC-004", "Reconnaissance Burst", types.SeverityMedium, score, priority,
				fmt.Sprintf("host:%s", host), "", "", host, list[len(list)-1].DetectedAt,
				len(list), int(window.Seconds()), []string{"T1087", "T1082", "T1059"},
				[]string{"Recon_Process_Chain"}, list)

			incidents = append(incidents, &domain.Incident{
				ID:             uuid.New(),
				OrganizationID: orgID,
				RuleID:         "INC-004",
				RuleName:       "Reconnaissance Burst",
				Category:       domain.RuleCategoryDiscovery,
				Severity:       types.SeverityMedium,
				Score:          score,
				Priority:       priority,
				EntityKey:      fmt.Sprintf("host:%s", host),
				Status:         domain.IncidentStatusNew,
				Title:          fmt.Sprintf("Reconnaissance Tool Burst on '%s'", host),
				Summary:        fmt.Sprintf("Host %s triggered %d discovery tools within 5 minutes. Host placed on 4-hour correlation watchlist.", host, len(list)),
				Evidence:       ev,
				OccurredAt:     list[len(list)-1].DetectedAt,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}
	}

	return incidents
}

// INC-005: Defense Impairment & Pre-Ransomware Prep
// >= 1 Defense Disabled AND >= 1 Shadow Copy Destruction on same host within 20m.
func (e *CorrelationEngine) evaluateINC005(orgID uuid.UUID) []*domain.Incident {
	window := 20 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	defDisabled := make(map[string][]*domain.Alert)
	shadowDelete := make(map[string][]*domain.Alert)

	for _, a := range alerts {
		if a.EntityHost == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Defense_Disabled_Defender", "Defense_Disabled_Service_Stopped", "Defense_Disabled_PowerShell":
			defDisabled[a.EntityHost] = append(defDisabled[a.EntityHost], a)
		case "Shadow_Copy_Destruction":
			shadowDelete[a.EntityHost] = append(shadowDelete[a.EntityHost], a)
		}
	}

	var incidents []*domain.Incident
	for host, defs := range defDisabled {
		shadows, ok := shadowDelete[host]
		if !ok || len(shadows) == 0 {
			continue
		}

		score := 85
		if e.store.IsHostOnWatchlist(host) {
			score += 10
		}
		if score > 100 {
			score = 100
		}
		priority := "P1"

		var contributingAlerts []*domain.Alert
		contributingAlerts = append(contributingAlerts, defs...)
		contributingAlerts = append(contributingAlerts, shadows...)

		ev := buildEvidence("INC-005", "Defense Impairment & Pre-Ransomware Prep", types.SeverityCritical, score, priority,
			fmt.Sprintf("host:%s", host), "", "", host, shadows[0].DetectedAt,
			len(contributingAlerts), int(window.Seconds()), []string{"T1562.001", "T1490"},
			[]string{"Defense_Disabled", "Shadow_Copy_Destruction"}, contributingAlerts)

		incidents = append(incidents, &domain.Incident{
			ID:             uuid.New(),
			OrganizationID: orgID,
			RuleID:         "INC-005",
			RuleName:       "Defense Impairment & Pre-Ransomware Prep",
			Category:       domain.RuleCategoryDefenseEvasion,
			Severity:       types.SeverityCritical,
			Score:          score,
			Priority:       priority,
			EntityKey:      fmt.Sprintf("host:%s", host),
			Status:         domain.IncidentStatusNew,
			Title:          fmt.Sprintf("Ransomware Preparation / Defense Impairment on '%s'", host),
			Summary:        fmt.Sprintf("Host %s suffered security defense disabling accompanied by shadow copy destruction.", host),
			Evidence:       ev,
			OccurredAt:     shadows[0].DetectedAt,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	return incidents
}

// INC-006: Credential Theft / Mimikatz
// >= 2 of {Mimikatz dropped, PS Mimikatz command, AV 1116->1119 failed remediation} on same host in 15m.
func (e *CorrelationEngine) evaluateINC006(orgID uuid.UUID) []*domain.Incident {
	window := 15 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	credTheftByHost := make(map[string][]*domain.Alert)
	for _, a := range alerts {
		if a.EntityHost == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Mimikatz_Dropped", "PowerShell_Mimikatz_Command", "Defender_Remediation_Failed":
			credTheftByHost[a.EntityHost] = append(credTheftByHost[a.EntityHost], a)
		}
	}

	var incidents []*domain.Incident
	for host, list := range credTheftByHost {
		if len(list) >= 2 {
			score := 85
			if e.store.IsHostOnWatchlist(host) {
				score += 10
			}
			if score > 100 {
				score = 100
			}
			priority := "P1"

			ev := buildEvidence("INC-006", "Credential Theft (Mimikatz Activity)", types.SeverityCritical, score, priority,
				fmt.Sprintf("host:%s", host), "", "", host, list[len(list)-1].DetectedAt,
				len(list), int(window.Seconds()), []string{"T1003.001", "T1562.001"},
				[]string{"Mimikatz_Activity"}, list)

			incidents = append(incidents, &domain.Incident{
				ID:             uuid.New(),
				OrganizationID: orgID,
				RuleID:         "INC-006",
				RuleName:       "Credential Theft (Mimikatz Activity)",
				Category:       domain.RuleCategoryCredentialAccess,
				Severity:       types.SeverityCritical,
				Score:          score,
				Priority:       priority,
				EntityKey:      fmt.Sprintf("host:%s", host),
				Status:         domain.IncidentStatusNew,
				Title:          fmt.Sprintf("Mimikatz Credential Theft Activity on '%s'", host),
				Summary:        fmt.Sprintf("Host %s triggered multiple credential dumping signals (Mimikatz execution or remediation failure).", host),
				Evidence:       ev,
				OccurredAt:     list[len(list)-1].DetectedAt,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}
	}

	return incidents
}

// INC-007: Persistence Mechanism Established
// >= 2 distinct persistence techniques on same host within 60m (RunKey, Startup LNK, IFEO, Service).
func (e *CorrelationEngine) evaluateINC007(orgID uuid.UUID) []*domain.Incident {
	window := 60 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	persistByHost := make(map[string]map[string]*domain.Alert)
	for _, a := range alerts {
		if a.EntityHost == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Persistence_RunKey", "Persistence_Startup_LNK", "Persistence_StickyKeys_IFEO", "Persistence_Malicious_Service":
			if _, ok := persistByHost[a.EntityHost]; !ok {
				persistByHost[a.EntityHost] = make(map[string]*domain.Alert)
			}
			persistByHost[a.EntityHost][a.ThreatLabel] = a
		}
	}

	var incidents []*domain.Incident
	for host, techMap := range persistByHost {
		if len(techMap) >= 2 {
			score := 65
			if e.store.IsHostOnWatchlist(host) {
				score += 10
			}
			if score > 100 {
				score = 100
			}
			priority := "P2"
			if score >= 90 {
				priority = "P1"
			}

			var distinctAlerts []*domain.Alert
			var techNames []string
			for name, alert := range techMap {
				distinctAlerts = append(distinctAlerts, alert)
				techNames = append(techNames, name)
			}

			ev := buildEvidence("INC-007", "Persistence Mechanism Established", types.SeverityHigh, score, priority,
				fmt.Sprintf("host:%s", host), "", "", host, distinctAlerts[0].DetectedAt,
				len(distinctAlerts), int(window.Seconds()), []string{"T1547.001", "T1546.008", "T1543.003"},
				techNames, distinctAlerts)

			incidents = append(incidents, &domain.Incident{
				ID:             uuid.New(),
				OrganizationID: orgID,
				RuleID:         "INC-007",
				RuleName:       "Persistence Mechanism Established",
				Category:       domain.RuleCategoryPersistence,
				Severity:       types.SeverityHigh,
				Score:          score,
				Priority:       priority,
				EntityKey:      fmt.Sprintf("host:%s", host),
				Status:         domain.IncidentStatusNew,
				Title:          fmt.Sprintf("Multiple Persistence Mechanisms Established on '%s'", host),
				Summary:        fmt.Sprintf("Host %s established %d distinct persistence techniques (%v).", host, len(techMap), techNames),
				Evidence:       ev,
				OccurredAt:     distinctAlerts[0].DetectedAt,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}
	}

	return incidents
}

// INC-008: C2 Channel Established
// 1 Outbound C2 signal AND 1 LOLBin signal on same host within 10m.
func (e *CorrelationEngine) evaluateINC008(orgID uuid.UUID) []*domain.Incident {
	window := 10 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	c2ByHost := make(map[string][]*domain.Alert)
	lolbinByHost := make(map[string][]*domain.Alert)

	for _, a := range alerts {
		if a.EntityHost == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Outbound_C2_Connection", "DynDNS_C2_Resolution", "Reverse_Shell_Connection":
			c2ByHost[a.EntityHost] = append(c2ByHost[a.EntityHost], a)
		case "LOLBin_Execution":
			lolbinByHost[a.EntityHost] = append(lolbinByHost[a.EntityHost], a)
		}
	}

	var incidents []*domain.Incident
	for host, c2s := range c2ByHost {
		lolbins, ok := lolbinByHost[host]
		if !ok || len(lolbins) == 0 {
			continue
		}

		score := 85
		if e.store.IsHostOnWatchlist(host) {
			score += 10
		}
		if score > 100 {
			score = 100
		}
		priority := "P1"

		var contributingAlerts []*domain.Alert
		contributingAlerts = append(contributingAlerts, c2s...)
		contributingAlerts = append(contributingAlerts, lolbins...)

		ev := buildEvidence("INC-008", "C2 Channel Established", types.SeverityCritical, score, priority,
			fmt.Sprintf("host:%s", host), "", "", host, c2s[0].DetectedAt,
			len(contributingAlerts), int(window.Seconds()), []string{"T1071.001", "T1059"},
			[]string{"C2_Outbound", "LOLBin_Execution"}, contributingAlerts)

		incidents = append(incidents, &domain.Incident{
			ID:             uuid.New(),
			OrganizationID: orgID,
			RuleID:         "INC-008",
			RuleName:       "C2 Channel Established",
			Category:       domain.RuleCategoryInitialAccess,
			Severity:       types.SeverityCritical,
			Score:          score,
			Priority:       priority,
			EntityKey:      fmt.Sprintf("host:%s", host),
			Status:         domain.IncidentStatusNew,
			Title:          fmt.Sprintf("Active Command & Control (C2) Established on '%s'", host),
			Summary:        fmt.Sprintf("Host %s established outbound C2 connection correlated with LOLBin execution.", host),
			Evidence:       ev,
			OccurredAt:     c2s[0].DetectedAt,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	return incidents
}

// INC-009: Data Exfiltration
// Strictly sequential: Collect -> Archive -> Upload within 30m on same host.
func (e *CorrelationEngine) evaluateINC009(orgID uuid.UUID) []*domain.Incident {
	window := 30 * time.Minute
	alerts := e.store.GetAlertsInWindow(time.Now().Add(-window))

	collectByHost := make(map[string][]*domain.Alert)
	archiveByHost := make(map[string][]*domain.Alert)
	uploadByHost := make(map[string][]*domain.Alert)

	for _, a := range alerts {
		if a.EntityHost == "" {
			continue
		}
		switch a.ThreatLabel {
		case "Data_Collection_Files":
			collectByHost[a.EntityHost] = append(collectByHost[a.EntityHost], a)
		case "Data_Archive_Compression":
			archiveByHost[a.EntityHost] = append(archiveByHost[a.EntityHost], a)
		case "Data_Exfiltration_Upload":
			uploadByHost[a.EntityHost] = append(uploadByHost[a.EntityHost], a)
		}
	}

	var incidents []*domain.Incident
	for host, collects := range collectByHost {
		archives, hasArchive := archiveByHost[host]
		uploads, hasUpload := uploadByHost[host]
		if !hasArchive || !hasUpload {
			continue
		}

		// Verify strict sequence: c.DetectedAt <= a.DetectedAt <= u.DetectedAt
		var matchedCollect, matchedArchive, matchedUpload *domain.Alert
		for _, c := range collects {
			for _, a := range archives {
				if a.DetectedAt.After(c.DetectedAt) || a.DetectedAt.Equal(c.DetectedAt) {
					for _, u := range uploads {
						if u.DetectedAt.After(a.DetectedAt) || u.DetectedAt.Equal(a.DetectedAt) {
							matchedCollect = c
							matchedArchive = a
							matchedUpload = u
							break
						}
					}
				}
				if matchedUpload != nil {
					break
				}
			}
			if matchedUpload != nil {
				break
			}
		}

		if matchedUpload == nil {
			continue
		}

		score := 85
		if e.store.IsHostOnWatchlist(host) {
			score += 10
		}
		if score > 100 {
			score = 100
		}
		priority := "P1"

		contributingAlerts := []*domain.Alert{matchedCollect, matchedArchive, matchedUpload}
		ev := buildEvidence("INC-009", "Data Exfiltration Pipeline", types.SeverityCritical, score, priority,
			fmt.Sprintf("host:%s", host), "", "", host, matchedUpload.DetectedAt,
			3, int(window.Seconds()), []string{"T1005", "T1560.001", "T1048"},
			[]string{"Data_Collection_Files", "Data_Archive_Compression", "Data_Exfiltration_Upload"}, contributingAlerts)

		incidents = append(incidents, &domain.Incident{
			ID:             uuid.New(),
			OrganizationID: orgID,
			RuleID:         "INC-009",
			RuleName:       "Data Exfiltration Pipeline",
			Category:       domain.RuleCategoryExfiltration,
			Severity:       types.SeverityCritical,
			Score:          score,
			Priority:       priority,
			EntityKey:      fmt.Sprintf("host:%s", host),
			Status:         domain.IncidentStatusNew,
			Title:          fmt.Sprintf("Data Exfiltration Pipeline Observed on '%s'", host),
			Summary:        fmt.Sprintf("Host %s performed sequential file collection, archive compression, and external upload within %v.", host, matchedUpload.DetectedAt.Sub(matchedCollect.DetectedAt)),
			Evidence:       ev,
			OccurredAt:     matchedUpload.DetectedAt,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	return incidents
}

// buildEvidence constructs a populated domain.Evidence object with MITRE, entity keys, and summaries.
func buildEvidence(ruleID, ruleName string, sev types.Severity, score int, priority, entityKey, targetUser, sourceIP, host string,
	occurredAt time.Time, attemptCount, windowSec int, mitre, threats []string, alerts ...interface{}) domain.Evidence {

	var summaries []domain.ContributingEventSummary
	for _, item := range alerts {
		switch v := item.(type) {
		case []*domain.Alert:
			for _, a := range v {
				summaries = append(summaries, alertToSummary(a))
			}
		case *domain.Alert:
			summaries = append(summaries, alertToSummary(v))
		}
	}

	return domain.Evidence{
		RuleID:              ruleID,
		RuleName:            ruleName,
		Severity:            sev,
		Score:               score,
		Priority:            priority,
		EntityKey:           entityKey,
		TargetUser:          targetUser,
		SourceIP:            sourceIP,
		HostName:            host,
		OccurredAt:          occurredAt,
		AttemptCount:        attemptCount,
		TimeWindowSeconds:   windowSec,
		MITRETechniques:     mitre,
		ContributingThreats: threats,
		ContributingEvents:  summaries,
	}
}

func alertToSummary(a *domain.Alert) domain.ContributingEventSummary {
	var raw map[string]any
	if a.RawEvent != nil {
		raw = a.RawEvent.RawPayload
	}
	return domain.ContributingEventSummary{
		EventID:       a.ID.String(),
		SourceEventID: a.EventID,
		EventType:     a.ThreatLabel,
		OccurredAt:    a.DetectedAt,
		ActorUsername: a.EntityAccount,
		IPAddress:     a.EntityIP,
		Summary:       fmt.Sprintf("[%s] %s on %s", a.LogSource, a.ThreatLabel, a.EntityHost),
		RawPayload:    raw,
	}
}

