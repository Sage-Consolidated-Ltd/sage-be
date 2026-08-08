package dto

// UpdateIdentityRequest updates user's identity info
type UpdateIdentityRequest struct {
	FullName    string  `json:"full_name,omitempty"`
	PhoneNumber string  `json:"phone_number,omitempty"`
	BackupEmail string  `json:"backup_email,omitempty" validate:"omitempty,email"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// UpdatePreferencesRequest updates UI and product behavior preferences
type UpdatePreferencesRequest struct {
	Theme                string `json:"theme,omitempty" validate:"omitempty,oneof=light dark system auto"`
	Timezone             string `json:"timezone,omitempty"`
	Language             string `json:"language,omitempty"`
	DateFormat           string `json:"date_format,omitempty"`
	DashboardDefaultView string `json:"dashboard_default_view,omitempty"`
	TablePageSize        int    `json:"table_page_size,omitempty"`
	AutoRefreshInterval  int    `json:"auto_refresh_interval,omitempty"`
}

// UpdateNotificationsRequest updates notification preferences
type UpdateNotificationsRequest struct {
	EmailEnabled              bool   `json:"email_enabled"`
	PushEnabled               bool   `json:"push_enabled"`
	SlackEnabled              bool   `json:"slack_enabled"`
	ProductUpdates            bool   `json:"product_updates"`
	WeeklySummary             bool   `json:"weekly_summary"`
	AlertSeverityThreshold    string `json:"alert_severity_threshold,omitempty" validate:"omitempty,oneof=low medium high critical"`
	NotifyOnNewAlert          bool   `json:"notify_on_new_alert"`
	NotifyOnIncidentUpdate    bool   `json:"notify_on_incident_update"`
	NotifyOnPlaybookExecution bool   `json:"notify_on_playbook_execution"`
}

// ChangePasswordRequest updates user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=12"`
}

// ConfigureBackupEmailRequest sets user backup email
type ConfigureBackupEmailRequest struct {
	BackupEmail string `json:"backup_email" validate:"required,email"`
}

// DeleteAccountRequest handles user account deletion confirmation
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" validate:"required"`
}

// UpdateProfileRequest is kept for backwards compatibility - updates basic profile fields
type UpdateProfileRequest struct {
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	TimeZone  *string `json:"time_zone,omitempty"`
}

// ListActivityRequest query params for activity list
type ListActivityRequest struct {
	Page       int    `json:"page" query:"page"`
	PageSize   int    `json:"page_size" query:"page_size"`
	StartTime  string `json:"start_time" query:"start_time"`
	EndTime    string `json:"end_time" query:"end_time"`
	ActionType string `json:"action_type" query:"action_type"`
}
