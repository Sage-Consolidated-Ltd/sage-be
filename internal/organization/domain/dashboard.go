package domain

import (
	"time"

	"github.com/google/uuid"
)

type DashboardTab string

const (
	TabOverview DashboardTab = "overview"
	TabAssets   DashboardTab = "assets"
	TabHealth   DashboardTab = "health"
	TabIdentity DashboardTab = "identity"
)

// OrganizationDashboard represents the unified, materialized security and operational dashboard for an organization.
type OrganizationDashboard struct {
	OrganizationID   uuid.UUID                 `json:"organization_id"`
	Tab              DashboardTab              `json:"tab,omitempty"`
	SecurityScore    *SecurityScore            `json:"security_score,omitempty"`
	Vulnerabilities  *VulnerabilitiesSummary   `json:"vulnerabilities,omitempty"`
	IdentityHealth   *IdentityHealthSummary    `json:"identity_health,omitempty"`
	EndpointCoverage *AssetProtectionCoverage  `json:"endpoint_coverage,omitempty"`
	ThreatIntel      *ThreatIntelFeedsSummary  `json:"threat_intel,omitempty"`
	ActiveIncidents  []ActiveIncident          `json:"active_incidents,omitempty"`
	DangerousThreats *AssetRiskDistribution    `json:"dangerous_threats,omitempty"`
	ComplianceRisks  *ComplianceRiskIndicators `json:"compliance_risks,omitempty"`
	ThreatTrends     *ThreatTrendsSummary      `json:"threat_trends,omitempty"`
	GeoThreats       *GeoThreatsSummary        `json:"geo_threats,omitempty"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

// Widget 1: Overall Security Posture Score
type SecurityScore struct {
	OverallScore           int                    `json:"overall_score"`
	WeeklyDelta            int                    `json:"weekly_delta"`
	Description            string                 `json:"description"`
	PendingRecommendations int                    `json:"pending_recommendations"`
	Pillars                SecurityPosturePillars `json:"pillars"`
}

type SecurityPosturePillars struct {
	ConfigHealth      int `json:"config_health"`
	Vulnerabilities   int `json:"vulnerabilities"`
	ThreatCoverage    int `json:"threat_coverage"`
	ResponseReadiness int `json:"response_readiness"`
}

// Widget 2: Known Vulnerabilities Breakdown
type VulnerabilitiesSummary struct {
	Critical     int `json:"critical"`
	High         int `json:"high"`
	Medium       int `json:"medium"`
	Low          int `json:"low"`
	Total        int `json:"total"`
	NewLast7Days int `json:"new_last_7_days"`
}

// Widget 3: Identity & Access Health
type IdentityHealthSummary struct {
	CoveragePercentage int `json:"coverage_percentage"`
	AccountsWithoutMFA int `json:"accounts_without_mfa"`
	DormantAccounts    int `json:"dormant_accounts"`
	ElevatedPrivileges int `json:"elevated_privileges"`
}

// Widget 4: Endpoint Protection Coverage
type AssetProtectionCoverage struct {
	CoveragePercentage int   `json:"coverage_percentage"`
	ProtectedCount     int64 `json:"protected_count"`
	UnprotectedCount   int64 `json:"unprotected_count"`
}

// Widget 5: Threat Intelligence Feeds
type ThreatIntelFeedsSummary struct {
	IndicatorsProcessed24h int64 `json:"indicators_processed_24h"`
	ActiveFeeds            int   `json:"active_feeds"`
	InactiveFeeds          int   `json:"inactive_feeds"`
}

// Widget 6: Active Incident Item
type ActiveIncident struct {
	ID           string    `json:"id"`
	IncidentName string    `json:"incident_name"`
	Severity     string    `json:"severity"`
	LastActivity time.Time `json:"last_activity"`
	Status       string    `json:"status"`
}

// Widget 7: Dangerous Threats Risk Distribution
type AssetRiskDistribution struct {
	OverallRiskPercentage int             `json:"overall_risk_percentage"`
	Breakdown             []AssetRiskItem `json:"breakdown"`
}

type AssetRiskItem struct {
	AssetName  string `json:"asset_name"`
	Percentage int    `json:"percentage"`
}

// Widget 8: Compliance Risk Indicators / Recent Activity
type ComplianceRiskIndicators struct {
	EncryptionVulnerabilities int    `json:"encryption_vulnerabilities"`
	ExcessiveUserPermissions  int    `json:"excessive_user_permissions"`
	OverlyTrustedUsers        int    `json:"overly_trusted_users"`
	VulnerabilitiesEmail      int    `json:"vulnerabilities_email"`
	DormantAccounts           int    `json:"dormant_accounts"`
	PhysicalSecurity          int    `json:"physical_security"`
	UnencryptedDevices        int    `json:"unencrypted_devices"`
	DetectionActionResult     string `json:"detection_action_result"`
}

// Widget 9: Threat Severity Trends
type ThreatTrendsSummary struct {
	CurrentMonth   string           `json:"current_month"`
	PreviousMonth  string           `json:"previous_month,omitempty"`
	SeverityFilter string           `json:"severity_filter,omitempty"`
	Days           []ThreatDayTrend `json:"days"`
}

type ThreatDayTrend struct {
	Day               int `json:"day"`
	Critical          int `json:"critical"`
	High              int `json:"high"`
	Medium            int `json:"medium"`
	Low               int `json:"low"`
	Total             int `json:"total"`
	CurrentMonthCount int `json:"current_month_count"`
	LastMonthCount    int `json:"last_month_count"`
}

// Widget 10: Live Geo Threat Origins Map (Embedded)
type GeoThreatsSummary struct {
	TotalThreats      int64               `json:"total_threats"`
	HighThreatRegion  string              `json:"high_threat_region"`
	MostTargetedAsset string              `json:"most_targeted_asset"`
	TopTargetedAssets []TargetedAssetInfo `json:"top_targeted_assets,omitempty"`
	Origins           []GeoThreatOrigin   `json:"origins"`
}

type TargetedAssetInfo struct {
	Asset string `json:"asset"`
	Count int64  `json:"count"`
	Type  string `json:"type,omitempty"`
}

type GeoThreatOrigin struct {
	Country    string  `json:"country"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}
