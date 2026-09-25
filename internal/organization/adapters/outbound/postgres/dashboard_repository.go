package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/organization/domain"
	"sage-backend/internal/organization/ports/outbound"
	"sage-backend/internal/shared/db"

	"github.com/google/uuid"
)

type DashboardSnapshotRepository struct {
	db.Repository
}

func NewDashboardSnapshotRepository(database *db.DB) outbound.DashboardRepository {
	return &DashboardSnapshotRepository{
		Repository: db.NewRepository(database),
	}
}

// GetSnapshot retrieves the pre-computed materialized dashboard snapshot from the database.
func (r *DashboardSnapshotRepository) GetSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
	const q = `
		SELECT 
			organization_id,
			security_score,
			vulnerabilities,
			identity_health,
			endpoint_coverage,
			threat_intel,
			active_incidents,
			dangerous_threats,
			compliance_risks,
			threat_trends,
			geo_threats,
			updated_at
		FROM organization_dashboard_snapshots
		WHERE organization_id = $1
	`

	var (
		storedOrgID      uuid.UUID
		securityScoreRaw []byte
		vulnsRaw         []byte
		identityRaw      []byte
		endpointRaw      []byte
		threatIntelRaw   []byte
		incidentsRaw     []byte
		threatsRaw       []byte
		complianceRaw    []byte
		trendsRaw        []byte
		geoThreatsRaw    []byte
		updatedAt        time.Time
	)

	err := r.Executor(ctx).QueryRowxContext(ctx, q, orgID).Scan(
		&storedOrgID,
		&securityScoreRaw,
		&vulnsRaw,
		&identityRaw,
		&endpointRaw,
		&threatIntelRaw,
		&incidentsRaw,
		&threatsRaw,
		&complianceRaw,
		&trendsRaw,
		&geoThreatsRaw,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If not yet materialized, compute live, save, and return
			return r.ComputeSnapshot(ctx, orgID)
		}
		return nil, fmt.Errorf("query dashboard snapshot: %w", err)
	}

	dashboard := &domain.OrganizationDashboard{
		OrganizationID: storedOrgID,
		UpdatedAt:      updatedAt,
	}

	_ = json.Unmarshal(securityScoreRaw, &dashboard.SecurityScore)
	_ = json.Unmarshal(vulnsRaw, &dashboard.Vulnerabilities)
	_ = json.Unmarshal(identityRaw, &dashboard.IdentityHealth)
	_ = json.Unmarshal(endpointRaw, &dashboard.EndpointCoverage)
	_ = json.Unmarshal(threatIntelRaw, &dashboard.ThreatIntel)
	_ = json.Unmarshal(incidentsRaw, &dashboard.ActiveIncidents)
	_ = json.Unmarshal(threatsRaw, &dashboard.DangerousThreats)
	_ = json.Unmarshal(complianceRaw, &dashboard.ComplianceRisks)
	_ = json.Unmarshal(trendsRaw, &dashboard.ThreatTrends)
	_ = json.Unmarshal(geoThreatsRaw, &dashboard.GeoThreats)

	return dashboard, nil
}

// SaveSnapshot upserts a materialized dashboard snapshot into the database.
func (r *DashboardSnapshotRepository) SaveSnapshot(ctx context.Context, snapshot *domain.OrganizationDashboard) error {
	if snapshot == nil {
		return nil
	}

	securityScoreJSON, _ := json.Marshal(snapshot.SecurityScore)
	vulnsJSON, _ := json.Marshal(snapshot.Vulnerabilities)
	identityJSON, _ := json.Marshal(snapshot.IdentityHealth)
	endpointJSON, _ := json.Marshal(snapshot.EndpointCoverage)
	threatIntelJSON, _ := json.Marshal(snapshot.ThreatIntel)
	incidentsJSON, _ := json.Marshal(snapshot.ActiveIncidents)
	threatsJSON, _ := json.Marshal(snapshot.DangerousThreats)
	complianceJSON, _ := json.Marshal(snapshot.ComplianceRisks)
	trendsJSON, _ := json.Marshal(snapshot.ThreatTrends)
	geoThreatsJSON, _ := json.Marshal(snapshot.GeoThreats)

	const upsertQ = `
		INSERT INTO organization_dashboard_snapshots (
			organization_id, security_score, vulnerabilities, identity_health,
			endpoint_coverage, threat_intel, active_incidents, dangerous_threats,
			compliance_risks, threat_trends, geo_threats, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (organization_id) DO UPDATE SET
			security_score    = EXCLUDED.security_score,
			vulnerabilities   = EXCLUDED.vulnerabilities,
			identity_health   = EXCLUDED.identity_health,
			endpoint_coverage = EXCLUDED.endpoint_coverage,
			threat_intel      = EXCLUDED.threat_intel,
			active_incidents  = EXCLUDED.active_incidents,
			dangerous_threats = EXCLUDED.dangerous_threats,
			compliance_risks  = EXCLUDED.compliance_risks,
			threat_trends     = EXCLUDED.threat_trends,
			geo_threats       = EXCLUDED.geo_threats,
			updated_at        = NOW()
	`

	_, err := r.Executor(ctx).ExecContext(
		ctx,
		upsertQ,
		snapshot.OrganizationID,
		securityScoreJSON,
		vulnsJSON,
		identityJSON,
		endpointJSON,
		threatIntelJSON,
		incidentsJSON,
		threatsJSON,
		complianceJSON,
		trendsJSON,
		geoThreatsJSON,
	)
	if err != nil {
		return fmt.Errorf("upsert dashboard snapshot: %w", err)
	}

	return nil
}

// ComputeSnapshot executes optimized multi-aggregate queries to calculate the full 10 widgets and saves the result.
func (r *DashboardSnapshotRepository) ComputeSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
	// 1. Compute Vulnerabilities
	vulns := r.computeVulnerabilities(ctx, orgID)

	// 2. Compute Security Score
	score := r.computeSecurityScore(ctx, orgID, vulns)

	// 3. Compute Identity & Access Health
	identity := r.computeIdentityHealth(ctx, orgID)

	// 4. Compute Endpoint Coverage
	endpoint := r.computeEndpointCoverage(ctx, orgID)

	// 5. Compute Threat Intel Feeds
	intel := r.computeThreatIntel(ctx, orgID)

	// 6. Compute Active Incidents
	incidents := r.computeActiveIncidents(ctx, orgID)

	// 7. Compute Dangerous Threats Risk Distribution
	dangerousThreats := r.computeDangerousThreats(ctx, orgID)

	// 8. Compute Compliance Indicators
	compliance := r.computeComplianceIndicators(ctx, orgID)

	// 9. Compute Threat Trends (Current vs Previous Month)
	trends := r.computeThreatTrends(ctx, orgID)

	// 10. Compute Live Geo Threat Origins Map
	geo := r.computeGeoThreats(ctx, orgID)

	snapshot := &domain.OrganizationDashboard{
		OrganizationID:   orgID,
		SecurityScore:    score,
		Vulnerabilities:  vulns,
		IdentityHealth:   identity,
		EndpointCoverage: endpoint,
		ThreatIntel:      intel,
		ActiveIncidents:  incidents,
		DangerousThreats: dangerousThreats,
		ComplianceRisks:  compliance,
		ThreatTrends:     trends,
		GeoThreats:       geo,
		UpdatedAt:        time.Now().UTC(),
	}

	// Persist snapshot to database asynchronously/best-effort so future calls hit snapshot table
	_ = r.SaveSnapshot(ctx, snapshot)

	return snapshot, nil
}

func (r *DashboardSnapshotRepository) computeVulnerabilities(ctx context.Context, orgID uuid.UUID) *domain.VulnerabilitiesSummary {
	const q = `
		SELECT 
			COUNT(*) FILTER (WHERE LOWER(severity) = 'critical')::int AS critical,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'high')::int AS high,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'medium')::int AS medium,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'low')::int AS low,
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days')::int AS new_7d
		FROM (
			SELECT severity, created_at FROM threats WHERE organization_id = $1
			UNION ALL
			SELECT severity, occurred_at AS created_at FROM security_events WHERE organization_id = $1
		) combined
	`

	var s domain.VulnerabilitiesSummary
	err := r.Executor(ctx).QueryRowxContext(ctx, q, orgID).Scan(
		&s.Critical,
		&s.High,
		&s.Medium,
		&s.Low,
		&s.Total,
		&s.NewLast7Days,
	)
	if err != nil {
		return &domain.VulnerabilitiesSummary{}
	}
	return &s
}

func (r *DashboardSnapshotRepository) computeSecurityScore(ctx context.Context, orgID uuid.UUID, vulns *domain.VulnerabilitiesSummary) *domain.SecurityScore {
	configHealth := 100
	var dqScore sql.NullInt32
	const dqQuery = `
		SELECT quality_score FROM data_quality_scans 
		WHERE organization_id = $1 AND status = 'completed' AND quality_score IS NOT NULL
		ORDER BY created_at DESC LIMIT 1
	`
	if err := r.Executor(ctx).QueryRowxContext(ctx, dqQuery, orgID).Scan(&dqScore); err == nil && dqScore.Valid {
		configHealth = int(dqScore.Int32)
	} else {
		var totalSources, activeSources int
		const dsQuery = `
			SELECT 
				COUNT(*)::int AS total,
				COUNT(*) FILTER (WHERE status = 'active')::int AS active
			FROM data_sources WHERE organization_id = $1 AND deleted_at IS NULL
		`
		if err := r.Executor(ctx).QueryRowxContext(ctx, dsQuery, orgID).Scan(&totalSources, &activeSources); err == nil && totalSources > 0 {
			configHealth = (activeSources * 100) / totalSources
		}
	}

	vulnScore := 100
	if vulns != nil {
		penalty := (vulns.Critical * 15) + (vulns.High * 8) + (vulns.Medium * 3) + (vulns.Low * 1)
		vulnScore -= penalty
		if vulnScore < 0 {
			vulnScore = 0
		}
	}

	threatCoverage := 100
	var totalDS, activeDS int
	const tcQuery = `
		SELECT 
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE status = 'active')::int AS active
		FROM data_sources WHERE organization_id = $1 AND deleted_at IS NULL
	`
	if err := r.Executor(ctx).QueryRowxContext(ctx, tcQuery, orgID).Scan(&totalDS, &activeDS); err == nil {
		if totalDS > 0 {
			threatCoverage = (activeDS * 100) / totalDS
		} else {
			threatCoverage = 0
		}
	}

	responseReadiness := 100
	var totalInc, resolvedInc int
	const incQuery = `
		SELECT 
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE LOWER(status) IN ('resolved', 'closed'))::int AS resolved
		FROM incidents WHERE organization_id = $1
	`
	if err := r.Executor(ctx).QueryRowxContext(ctx, incQuery, orgID).Scan(&totalInc, &resolvedInc); err == nil && totalInc > 0 {
		responseReadiness = (resolvedInc * 100) / totalInc
	}

	overallScore := (configHealth*20 + vulnScore*35 + threatCoverage*25 + responseReadiness*20) / 100
	if overallScore < 0 {
		overallScore = 0
	} else if overallScore > 100 {
		overallScore = 100
	}

	var pendingRecs int
	const recQuery = `
		SELECT COUNT(*)::int FROM threats 
		WHERE organization_id = $1 AND recommendation IS NOT NULL AND TRIM(recommendation) != ''
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, recQuery, orgID).Scan(&pendingRecs)

	return &domain.SecurityScore{
		OverallScore:           overallScore,
		WeeklyDelta:            0,
		Description:            "Security score is calculated dynamically from configuration health, vulnerabilities, threat coverage, and response readiness.",
		PendingRecommendations: pendingRecs,
		Pillars: domain.SecurityPosturePillars{
			ConfigHealth:      configHealth,
			Vulnerabilities:   vulnScore,
			ThreatCoverage:    threatCoverage,
			ResponseReadiness: responseReadiness,
		},
	}
}

func (r *DashboardSnapshotRepository) computeIdentityHealth(ctx context.Context, orgID uuid.UUID) *domain.IdentityHealthSummary {
	const q = `
		SELECT 
			COUNT(om.id)::int AS total_members,
			COUNT(om.id) FILTER (WHERE u.two_factor_enabled = true)::int AS mfa_enabled,
			COUNT(om.id) FILTER (WHERE u.two_factor_enabled = false)::int AS no_mfa,
			COUNT(om.id) FILTER (WHERE om.status IN ('inactive', 'suspended'))::int AS dormant,
			COUNT(om.id) FILTER (WHERE LOWER(COALESCE(r.name, '')) IN ('owner', 'admin', 'super_admin'))::int AS elevated
		FROM organization_members om
		JOIN users u ON om.user_id = u.id
		LEFT JOIN organization_roles r ON om.role_id = r.id
		WHERE om.organization_id = $1 AND u.deleted_at IS NULL
	`

	var total, mfaEnabled, noMFA, dormant, elevated int
	err := r.Executor(ctx).QueryRowxContext(ctx, q, orgID).Scan(&total, &mfaEnabled, &noMFA, &dormant, &elevated)
	if err != nil {
		return &domain.IdentityHealthSummary{}
	}

	coverage := 0
	if total > 0 {
		coverage = (mfaEnabled * 100) / total
	}

	return &domain.IdentityHealthSummary{
		CoveragePercentage: coverage,
		AccountsWithoutMFA: noMFA,
		DormantAccounts:    dormant,
		ElevatedPrivileges: elevated,
	}
}

func (r *DashboardSnapshotRepository) computeEndpointCoverage(ctx context.Context, orgID uuid.UUID) *domain.AssetProtectionCoverage {
	const q = `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'active' AND deleted_at IS NULL)::bigint AS protected,
			COUNT(*) FILTER (WHERE status != 'active' AND deleted_at IS NULL)::bigint AS unprotected,
			COUNT(*) FILTER (WHERE deleted_at IS NULL)::bigint AS total
		FROM data_sources 
		WHERE organization_id = $1
	`
	var protected, unprotected, total int64
	err := r.Executor(ctx).QueryRowxContext(ctx, q, orgID).Scan(&protected, &unprotected, &total)
	if err != nil {
		return &domain.AssetProtectionCoverage{}
	}

	coverage := 0
	if total > 0 {
		coverage = int((protected * 100) / total)
	}

	return &domain.AssetProtectionCoverage{
		CoveragePercentage: coverage,
		ProtectedCount:     protected,
		UnprotectedCount:   unprotected,
	}
}

func (r *DashboardSnapshotRepository) computeThreatIntel(ctx context.Context, orgID uuid.UUID) *domain.ThreatIntelFeedsSummary {
	const q = `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'active' AND deleted_at IS NULL)::int AS active,
			COUNT(*) FILTER (WHERE status != 'active' AND deleted_at IS NULL)::int AS inactive,
			COALESCE(SUM(events_today), 0)::bigint AS processed_today
		FROM data_sources
		WHERE organization_id = $1
	`
	var active, inactive int
	var processedToday int64
	err := r.Executor(ctx).QueryRowxContext(ctx, q, orgID).Scan(&active, &inactive, &processedToday)
	if err != nil {
		return &domain.ThreatIntelFeedsSummary{}
	}

	if processedToday == 0 {
		const evQuery = `
			SELECT COUNT(*)::bigint 
			FROM security_events 
			WHERE organization_id = $1 AND ingested_at >= NOW() - INTERVAL '24 hours'
		`
		_ = r.Executor(ctx).QueryRowxContext(ctx, evQuery, orgID).Scan(&processedToday)
	}

	return &domain.ThreatIntelFeedsSummary{
		IndicatorsProcessed24h: processedToday,
		ActiveFeeds:            active,
		InactiveFeeds:          inactive,
	}
}

func (r *DashboardSnapshotRepository) computeActiveIncidents(ctx context.Context, orgID uuid.UUID) []domain.ActiveIncident {
	const q = `
		SELECT 
			id::text,
			title,
			severity,
			occurred_at,
			status
		FROM incidents
		WHERE organization_id = $1 AND LOWER(status) NOT IN ('resolved', 'closed')
		ORDER BY occurred_at DESC
		LIMIT 5
	`

	rows, err := r.Executor(ctx).QueryContext(ctx, q, orgID)
	if err == nil {
		defer rows.Close()
		var incidents []domain.ActiveIncident
		for rows.Next() {
			var inc domain.ActiveIncident
			if err := rows.Scan(&inc.ID, &inc.IncidentName, &inc.Severity, &inc.LastActivity, &inc.Status); err == nil {
				incidents = append(incidents, inc)
			}
		}
		if len(incidents) > 0 {
			return incidents
		}
	}

	const alertQ = `
		SELECT 
			id::text,
			threat_label,
			'High' AS severity,
			detected_at,
			'Active' AS status
		FROM alerts
		WHERE organization_id = $1
		ORDER BY detected_at DESC
		LIMIT 5
	`
	alertRows, err := r.Executor(ctx).QueryContext(ctx, alertQ, orgID)
	if err == nil {
		defer alertRows.Close()
		var incidents []domain.ActiveIncident
		for alertRows.Next() {
			var inc domain.ActiveIncident
			if err := alertRows.Scan(&inc.ID, &inc.IncidentName, &inc.Severity, &inc.LastActivity, &inc.Status); err == nil {
				incidents = append(incidents, inc)
			}
		}
		if len(incidents) > 0 {
			return incidents
		}
	}

	return make([]domain.ActiveIncident, 0)
}

func (r *DashboardSnapshotRepository) computeDangerousThreats(ctx context.Context, orgID uuid.UUID) *domain.AssetRiskDistribution {
	const q = `
		SELECT 
			COALESCE(NULLIF(TRIM(entity_host), ''), 'unassigned-host') AS asset,
			COUNT(*)::int AS cnt
		FROM alerts
		WHERE organization_id = $1 AND entity_host IS NOT NULL AND TRIM(entity_host) != ''
		GROUP BY entity_host
		ORDER BY cnt DESC
		LIMIT 5
	`
	rows, err := r.Executor(ctx).QueryContext(ctx, q, orgID)
	if err == nil {
		defer rows.Close()
		var items []domain.AssetRiskItem
		totalCnt := 0
		for rows.Next() {
			var asset string
			var cnt int
			if err := rows.Scan(&asset, &cnt); err == nil {
				items = append(items, domain.AssetRiskItem{AssetName: asset, Percentage: cnt})
				totalCnt += cnt
			}
		}

		if totalCnt > 0 {
			for i := range items {
				items[i].Percentage = int((float64(items[i].Percentage) / float64(totalCnt)) * 100)
			}
			overallRisk := 100
			if totalCnt < 10 {
				overallRisk = totalCnt * 10
			}
			return &domain.AssetRiskDistribution{
				OverallRiskPercentage: overallRisk,
				Breakdown:             items,
			}
		}
	}

	const dsRiskQ = `
		SELECT 
			COALESCE(ds.name, 'unnamed-source') AS asset,
			COUNT(*)::int AS cnt
		FROM security_events se
		JOIN data_sources ds ON se.source_id = ds.id
		WHERE se.organization_id = $1 AND se.severity IN ('critical', 'high')
		GROUP BY ds.name
		ORDER BY cnt DESC
		LIMIT 5
	`
	dsRows, err := r.Executor(ctx).QueryContext(ctx, dsRiskQ, orgID)
	if err == nil {
		defer dsRows.Close()
		var items []domain.AssetRiskItem
		totalCnt := 0
		for dsRows.Next() {
			var asset string
			var cnt int
			if err := dsRows.Scan(&asset, &cnt); err == nil {
				items = append(items, domain.AssetRiskItem{AssetName: asset, Percentage: cnt})
				totalCnt += cnt
			}
		}
		if totalCnt > 0 {
			for i := range items {
				items[i].Percentage = int((float64(items[i].Percentage) / float64(totalCnt)) * 100)
			}
			overallRisk := 100
			if totalCnt < 10 {
				overallRisk = totalCnt * 10
			}
			return &domain.AssetRiskDistribution{
				OverallRiskPercentage: overallRisk,
				Breakdown:             items,
			}
		}
	}

	return &domain.AssetRiskDistribution{
		OverallRiskPercentage: 0,
		Breakdown:             make([]domain.AssetRiskItem, 0),
	}
}

func (r *DashboardSnapshotRepository) computeComplianceIndicators(ctx context.Context, orgID uuid.UUID) *domain.ComplianceRiskIndicators {
	indicators := &domain.ComplianceRiskIndicators{
		DetectionActionResult: "No compliance violations detected across connected data sources.",
	}

	const dormantQ = `
		SELECT COUNT(*)::int 
		FROM organization_members 
		WHERE organization_id = $1 AND status IN ('inactive', 'suspended')
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, dormantQ, orgID).Scan(&indicators.DormantAccounts)

	const mfaRiskQ = `
		SELECT COUNT(om.id)::int 
		FROM organization_members om
		JOIN users u ON om.user_id = u.id
		LEFT JOIN organization_roles r ON om.role_id = r.id
		WHERE om.organization_id = $1 
		  AND LOWER(COALESCE(r.name, '')) IN ('owner', 'admin', 'super_admin') 
		  AND u.two_factor_enabled = false
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, mfaRiskQ, orgID).Scan(&indicators.ExcessiveUserPermissions)

	const trustedQ = `
		SELECT COUNT(om.id)::int 
		FROM organization_members om
		LEFT JOIN organization_roles r ON om.role_id = r.id
		WHERE om.organization_id = $1 
		  AND LOWER(COALESCE(r.name, '')) IN ('owner', 'admin', 'super_admin')
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, trustedQ, orgID).Scan(&indicators.OverlyTrustedUsers)

	const encQ = `
		SELECT COUNT(*)::int 
		FROM threats 
		WHERE organization_id = $1 
		  AND (LOWER(title) LIKE '%encrypt%' OR LOWER(category) LIKE '%encrypt%' OR LOWER(what_happened) LIKE '%encrypt%')
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, encQ, orgID).Scan(&indicators.EncryptionVulnerabilities)

	const emailQ = `
		SELECT COUNT(*)::int 
		FROM threats 
		WHERE organization_id = $1 
		  AND (LOWER(title) LIKE '%email%' OR LOWER(category) LIKE '%phishing%' OR LOWER(category) LIKE '%email%')
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, emailQ, orgID).Scan(&indicators.VulnerabilitiesEmail)

	const alertsRiskQ = `
		SELECT 
			COUNT(*) FILTER (WHERE LOWER(threat_label) LIKE '%physical%')::int AS physical,
			COUNT(*) FILTER (WHERE LOWER(threat_label) LIKE '%unencrypt%' OR LOWER(threat_label) LIKE '%device%')::int AS devices
		FROM alerts
		WHERE organization_id = $1
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, alertsRiskQ, orgID).Scan(&indicators.PhysicalSecurity, &indicators.UnencryptedDevices)

	var totalAlerts int
	const alertsTotalQ = `SELECT COUNT(*)::int FROM alerts WHERE organization_id = $1`
	if err := r.Executor(ctx).QueryRowxContext(ctx, alertsTotalQ, orgID).Scan(&totalAlerts); err == nil && totalAlerts > 0 {
		indicators.DetectionActionResult = fmt.Sprintf("%d security alerts and threats actively analyzed.", totalAlerts)
	}

	return indicators
}

func (r *DashboardSnapshotRepository) computeThreatTrends(ctx context.Context, orgID uuid.UUID) *domain.ThreatTrendsSummary {
	now := time.Now().UTC()
	currentMonthName := now.Format("January")
	prevMonthName := now.AddDate(0, -1, 0).Format("January")

	type dayCounts struct {
		critical int
		high     int
		medium   int
		low      int
		total    int
	}

	curMonthCounts := make(map[int]dayCounts)
	lastMonthCounts := make(map[int]int)

	const curMonthQ = `
		SELECT 
			EXTRACT(DAY FROM occurred_at)::int AS d,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'critical')::int AS critical,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'high')::int AS high,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'medium')::int AS medium,
			COUNT(*) FILTER (WHERE LOWER(severity) = 'low')::int AS low,
			COUNT(*)::int AS total
		FROM security_events
		WHERE organization_id = $1 AND occurred_at >= DATE_TRUNC('month', NOW())
		GROUP BY d
	`
	if rows, err := r.Executor(ctx).QueryContext(ctx, curMonthQ, orgID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var d, c, h, m, l, tot int
			if err := rows.Scan(&d, &c, &h, &m, &l, &tot); err == nil {
				curMonthCounts[d] = dayCounts{critical: c, high: h, medium: m, low: l, total: tot}
			}
		}
	}

	const lastMonthQ = `
		SELECT 
			EXTRACT(DAY FROM occurred_at)::int AS d,
			COUNT(*)::int AS total
		FROM security_events
		WHERE organization_id = $1 
		  AND occurred_at >= DATE_TRUNC('month', NOW() - INTERVAL '1 month')
		  AND occurred_at < DATE_TRUNC('month', NOW())
		GROUP BY d
	`
	if rows, err := r.Executor(ctx).QueryContext(ctx, lastMonthQ, orgID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var d, tot int
			if err := rows.Scan(&d, &tot); err == nil {
				lastMonthCounts[d] = tot
			}
		}
	}

	currentDay := now.Day()
	if currentDay < 1 {
		currentDay = 1
	}

	days := make([]domain.ThreatDayTrend, 0, currentDay)
	for i := 1; i <= currentDay; i++ {
		cur := curMonthCounts[i]
		last := lastMonthCounts[i]
		days = append(days, domain.ThreatDayTrend{
			Day:               i,
			Critical:          cur.critical,
			High:              cur.high,
			Medium:            cur.medium,
			Low:               cur.low,
			Total:             cur.total,
			CurrentMonthCount: cur.total,
			LastMonthCount:    last,
		})
	}

	return &domain.ThreatTrendsSummary{
		CurrentMonth:  currentMonthName,
		PreviousMonth: prevMonthName,
		Days:          days,
	}
}

func (r *DashboardSnapshotRepository) computeGeoThreats(ctx context.Context, orgID uuid.UUID) *domain.GeoThreatsSummary {
	const q = `
		WITH combined_origins AS (
			SELECT TRIM(geo_country) AS country
			FROM security_events
			WHERE organization_id = $1 
			  AND geo_country IS NOT NULL 
			  AND TRIM(geo_country) != ''

			UNION ALL

			SELECT TRIM(context->>'geo_country') AS country
			FROM alerts
			WHERE organization_id = $1
			  AND context->>'geo_country' IS NOT NULL
			  AND TRIM(context->>'geo_country') != ''

			UNION ALL

			SELECT TRIM(se.geo_country) AS country
			FROM threats t
			JOIN analysis_results ar ON ar.id = t.analysis_id
			JOIN security_events se  ON se.source_event_id = ar.log_file_id::text
			WHERE t.organization_id = $1
			  AND se.geo_country IS NOT NULL 
			  AND TRIM(se.geo_country) != ''
		)
		SELECT country, COUNT(*)::bigint AS count
		FROM combined_origins
		WHERE country != ''
		GROUP BY country
		ORDER BY count DESC
		LIMIT 15
	`

	type originRow struct {
		Country string
		Count   int64
	}

	rows, err := r.Executor(ctx).QueryContext(ctx, q, orgID)
	if err != nil {
		return &domain.GeoThreatsSummary{
			TotalThreats:      0,
			HighThreatRegion:  "None",
			MostTargetedAsset: "None",
			Origins:           make([]domain.GeoThreatOrigin, 0),
		}
	}
	defer rows.Close()

	var list []originRow
	var totalThreats int64
	for rows.Next() {
		var row originRow
		if err := rows.Scan(&row.Country, &row.Count); err == nil {
			list = append(list, row)
			totalThreats += row.Count
		}
	}

	if len(list) == 0 {
		return &domain.GeoThreatsSummary{
			TotalThreats:      0,
			HighThreatRegion:  "None",
			MostTargetedAsset: "None",
			Origins:           make([]domain.GeoThreatOrigin, 0),
		}
	}

	origins := make([]domain.GeoThreatOrigin, 0, len(list))
	for _, it := range list {
		lat, lng := domain.ResolveCountryCoordinates(it.Country)
		pct := 0.0
		if totalThreats > 0 {
			pct = (float64(it.Count) / float64(totalThreats)) * 100
		}
		origins = append(origins, domain.GeoThreatOrigin{
			Country:    it.Country,
			Lat:        lat,
			Lng:        lng,
			Count:      it.Count,
			Percentage: pct,
		})
	}

	highThreatRegion := origins[0].Country

	mostTargetedAsset := "None"
	const assetQ = `
		SELECT COALESCE(NULLIF(TRIM(entity_host), ''), '') AS asset
		FROM alerts
		WHERE organization_id = $1 AND entity_host IS NOT NULL AND TRIM(entity_host) != ''
		GROUP BY entity_host
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`
	_ = r.Executor(ctx).QueryRowxContext(ctx, assetQ, orgID).Scan(&mostTargetedAsset)
	if mostTargetedAsset == "None" || mostTargetedAsset == "" {
		const dsAssetQ = `
			SELECT name FROM data_sources 
			WHERE organization_id = $1 AND deleted_at IS NULL 
			ORDER BY total_events DESC LIMIT 1
		`
		_ = r.Executor(ctx).QueryRowxContext(ctx, dsAssetQ, orgID).Scan(&mostTargetedAsset)
		if mostTargetedAsset == "" {
			mostTargetedAsset = "None"
		}
	}

	return &domain.GeoThreatsSummary{
		TotalThreats:      totalThreats,
		HighThreatRegion:  strings.Title(highThreatRegion),
		MostTargetedAsset: mostTargetedAsset,
		Origins:           origins,
	}
}
