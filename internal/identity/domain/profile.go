package domain

import (
	"database/sql"
	"time"
)

type UserPreferences struct {
	ID                   string         `json:"id" db:"id"`
	UserID               string         `json:"user_id" db:"user_id"`
	OrganizationID       sql.NullString `json:"organization_id" db:"organization_id"`
	Theme                string         `json:"theme" db:"theme"`
	Language             string         `json:"language" db:"language"`
	Timezone             string         `json:"timezone" db:"timezone"`
	DateFormat           string         `json:"date_format" db:"date_format"`
	DashboardDefaultView string         `json:"dashboard_default_view" db:"dashboard_default_view"`
	TablePageSize        int            `json:"table_page_size" db:"table_page_size"`
	AutoRefreshInterval  int            `json:"auto_refresh_interval" db:"auto_refresh_interval"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

type UserNotifications struct {
	ID                        string         `json:"id" db:"id"`
	UserID                    string         `json:"user_id" db:"user_id"`
	OrganizationID            sql.NullString `json:"organization_id" db:"organization_id"`
	EmailEnabled              bool           `json:"email_enabled" db:"email_enabled"`
	SlackEnabled              bool           `json:"slack_enabled" db:"slack_enabled"`
	PushEnabled               bool           `json:"push_enabled" db:"push_enabled"`
	ProductUpdates            bool           `json:"product_updates" db:"product_updates"`
	WeeklySummary             bool           `json:"weekly_summary" db:"weekly_summary"`
	CriticalAlertsOnly        bool           `json:"critical_alerts_only" db:"critical_alerts_only"`
	WeeklyReportEnabled       bool           `json:"weekly_report_enabled" db:"weekly_report_enabled"`
	SecurityAlertsFormat      string         `json:"security_alerts_format" db:"security_alerts_format"`
	AlertSeverityThreshold    string         `json:"alert_severity_threshold" db:"alert_severity_threshold"`
	NotifyOnNewAlert          bool           `json:"notify_on_new_alert" db:"notify_on_new_alert"`
	NotifyOnIncidentUpdate    bool           `json:"notify_on_incident_update" db:"notify_on_incident_update"`
	NotifyOnPlaybookExecution bool           `json:"notify_on_playbook_execution" db:"notify_on_playbook_execution"`
	CreatedAt                 time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at" db:"updated_at"`
}

type AuditLog struct {
	ID             string         `json:"id" db:"id"`
	UserID         string         `json:"user_id" db:"user_id"`
	OrganizationID sql.NullString `json:"organization_id" db:"organization_id"`
	ActionType     string         `json:"action_type" db:"action_type"`
	ActionCategory string         `json:"action_category" db:"action_category"`
	ResourceType   sql.NullString `json:"resource_type" db:"resource_type"`
	ResourceID     sql.NullString `json:"resource_id" db:"resource_id"`
	ResourceTarget sql.NullString `json:"resource_target" db:"resource_target"`
	IPAddress      sql.NullString `json:"ip_address" db:"ip_address"`
	UserAgent      sql.NullString `json:"user_agent" db:"user_agent"`
	Metadata       sql.NullString `json:"metadata" db:"metadata"`
	Status         string         `json:"status" db:"status"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

type UserActivityResponse struct {
	Logs       []AuditLog `json:"logs"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}
