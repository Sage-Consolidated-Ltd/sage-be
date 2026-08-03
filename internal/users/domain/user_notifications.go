package domain

import (
	"database/sql"
	"time"
)

type UserNotifications struct {
	ID                        string         `json:"id" db:"id"`
	UserID                    string         `json:"user_id" db:"user_id"`
	OrganizationID            sql.NullString `json:"organization_id" db:"organization_id"`
	EmailEnabled              bool           `json:"email_enabled" db:"email_enabled"`
	PushEnabled               bool           `json:"push_enabled" db:"push_enabled"`
	SlackEnabled              bool           `json:"slack_enabled" db:"slack_enabled"`
	AlertSeverityThreshold    string         `json:"alert_severity_threshold" db:"alert_severity_threshold"`
	NotifyOnNewAlert          bool           `json:"notify_on_new_alert" db:"notify_on_new_alert"`
	NotifyOnIncidentUpdate    bool           `json:"notify_on_incident_update" db:"notify_on_incident_update"`
	NotifyOnPlaybookExecution bool           `json:"notify_on_playbook_execution" db:"notify_on_playbook_execution"`
	CreatedAt                 time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at" db:"updated_at"`
}

type UserNotificationsResponse struct {
	EmailEnabled              bool   `json:"email_enabled"`
	PushEnabled               bool   `json:"push_enabled"`
	SlackEnabled              bool   `json:"slack_enabled"`
	AlertSeverityThreshold    string `json:"alert_severity_threshold"`
	NotifyOnNewAlert          bool   `json:"notify_on_new_alert"`
	NotifyOnIncidentUpdate    bool   `json:"notify_on_incident_update"`
	NotifyOnPlaybookExecution bool   `json:"notify_on_playbook_execution"`
}

func (n *UserNotifications) ToResponse() *UserNotificationsResponse {
	return &UserNotificationsResponse{
		EmailEnabled:              n.EmailEnabled,
		PushEnabled:               n.PushEnabled,
		SlackEnabled:              n.SlackEnabled,
		AlertSeverityThreshold:    n.AlertSeverityThreshold,
		NotifyOnNewAlert:          n.NotifyOnNewAlert,
		NotifyOnIncidentUpdate:    n.NotifyOnIncidentUpdate,
		NotifyOnPlaybookExecution: n.NotifyOnPlaybookExecution,
	}
}
