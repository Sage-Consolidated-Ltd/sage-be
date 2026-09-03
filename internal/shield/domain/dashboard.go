package domain

import "time"

// Widget 1: Overall Security Posture Score
type SecurityPostureScore struct {
	OverallScore           int                    `json:"overall_score"`
	WeeklyDelta            int                    `json:"weekly_delta"`
	Description            string                 `json:"description"`
	PendingRecommendations int                    `json:"pending_recommendations"`
	Pillars                SecurityPosturePillars `json:"pillars"`
}

type SecurityPosturePillars struct {
	ConfigHealth     int `json:"config_health"`
	Vulnerabilities  int `json:"vulnerabilities"`
	ThreatCoverage   int `json:"threat_coverage"`
	ResponseReadiness int `json:"response_readiness"`
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

// Widget 7: Dangerous Threats Risk Distribution by Asset
type AssetRiskDistribution struct {
	OverallRiskPercentage int             `json:"overall_risk_percentage"`
	Breakdown             []AssetRiskItem `json:"breakdown"`
}

type AssetRiskItem struct {
	AssetName  string `json:"asset_name"`
	Percentage int    `json:"percentage"`
}

// Widget 8: Compliance Risk Indicators
type ComplianceRiskIndicators struct {
	EncryptionVulnerabilities int     `json:"encryption_vulnerabilities"`
	ExcessiveUserPermissions  int     `json:"excessive_user_permissions"`
	OverlyTrustedUsers        int     `json:"overly_trusted_users"`
	VulnerabilitiesEmail      int     `json:"vulnerabilities_email"`
	DormantAccounts           int     `json:"dormant_accounts"`
	PhysicalSecurity          int     `json:"physical_security"`
	UnencryptedDevices        int     `json:"unencrypted_devices"`
	DetectionActionResult     string  `json:"detection_action_result"`
}

// Widget 9: Threat Severity Trends Line Chart
type ThreatTrendsSummary struct {
	CurrentMonth  string           `json:"current_month"`
	PreviousMonth string           `json:"previous_month,omitempty"`
	Days          []ThreatDayTrend `json:"days"`
}

type ThreatDayTrend struct {
	Day              int `json:"day"`
	CurrentMonthCount int `json:"current_month_count"`
	LastMonthCount    int `json:"last_month_count"`
}

// Widget 10: Geo-IP Threat Map
type GeoThreatsSummary struct {
	TotalThreats       int64             `json:"total_threats"`
	HighThreatRegion   string            `json:"high_threat_region"`
	MostTargetedAsset  string            `json:"most_targeted_asset"`
	Origins            []GeoThreatOrigin `json:"origins"`
}

type GeoThreatOrigin struct {
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Count   int64   `json:"count"`
}
