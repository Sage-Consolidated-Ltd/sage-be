package http

import (
	"time"

	_ "sage-backend/internal/identity/usecase/dto"
	_ "sage-backend/internal/shared/response"
)

type UpdateProfileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UserPreferencesDoc struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	OrganizationID       *string   `json:"organization_id,omitempty"`
	Theme                string    `json:"theme"`
	Language             string    `json:"language"`
	Timezone             string    `json:"timezone"`
	DashboardDefaultView string    `json:"dashboard_default_view"`
	TablePageSize        int       `json:"table_page_size"`
	AutoRefreshInterval  int       `json:"auto_refresh_interval"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type UserNotificationsDoc struct {
	ID                        string    `json:"id"`
	UserID                    string    `json:"user_id"`
	OrganizationID            *string   `json:"organization_id,omitempty"`
	EmailEnabled              bool      `json:"email_enabled"`
	SlackEnabled              bool      `json:"slack_enabled"`
	PushEnabled               bool      `json:"push_enabled"`
	CriticalAlertsOnly        bool      `json:"critical_alerts_only"`
	WeeklyReportEnabled       bool      `json:"weekly_report_enabled"`
	SecurityAlertsFormat      string    `json:"security_alerts_format"`
	AlertSeverityThreshold    string    `json:"alert_severity_threshold"`
	NotifyOnNewAlert          bool      `json:"notify_on_new_alert"`
	NotifyOnIncidentUpdate    bool      `json:"notify_on_incident_update"`
	NotifyOnPlaybookExecution bool      `json:"notify_on_playbook_execution"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type AuditLogDoc struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	ActionType     string    `json:"action_type"`
	ActionCategory string    `json:"action_category"`
	ResourceType   *string   `json:"resource_type,omitempty"`
	ResourceID     *string   `json:"resource_id,omitempty"`
	ResourceTarget *string   `json:"resource_target,omitempty"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	Metadata       *string   `json:"metadata,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type GetActivityResponseData struct {
	Data     []AuditLogDoc `json:"data"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type GetActivityResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    GetActivityResponseData `json:"data"`
}

type UploadAvatarResponseData struct {
	AvatarURL string `json:"avatar_url"`
}

type UploadAvatarResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    UploadAvatarResponseData `json:"data"`
}

type UserUpdateIdentityRequest struct {
	FullName string `json:"full_name,omitempty"`
}

// @Summary Get User Profile
// @Description Retrieves the comprehensive identity profile of the currently authenticated user, including personal details and organizational information.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.ProfileResponse
// @Router /profile [get]
func _GetProfile() {}

// @Summary Update User Profile
// @Description Updates the user's profile information such as name, email, and other personal details.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param avatar formData file true "Avatar image file"
// @Param request body UserUpdateIdentityRequest true "Update Profile Request"
// @Success 200 {object} UpdateProfileResponse
// @Router /profile [patch]
func _UpdateProfile() {}

// @Summary Get User Preferences
// @Description Retrieves the user's preferences for notifications, theme, and other customizable settings.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} UserPreferencesDoc
// @Router /profile/preferences [get]
func _GetPreferences() {}

// @Summary Update User Preferences
// @Description Updates the user's preferences for notifications, theme, and other customizable settings.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdatePreferencesRequest true "Update Preferences Request"
// @Success 200 {object} response.Response
// @Router /profile/preferences [patch]
func _UpdatePreferences() {}

// @Summary Get User Notifications
// @Description Retrieves the user's notification settings and history.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} UserNotificationsDoc
// @Router /profile/notifications [get]
func _GetNotifications() {}

// @Summary Update User Notifications
// @Description Updates the user's notification settings.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.UpdateNotificationsRequest true "Update Notifications Request"
// @Success 200 {object} response.Response
// @Router /profile/notifications [patch]
func _UpdateNotifications() {}

// @Summary Get User Session
// @Description Retrieves details about the user's current session, including login time, IP address, and device information.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.UserSessionResponse
// @Router /profile/session [get]
func _GetSession() {}

// @Summary Revoke User Session
// @Description Revokes the user's current session, effectively logging them out from the current device.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Router /profile/session/revoke [post]
func _RevokeSession() {}

// @Summary Get User Activity
// @Description Retrieves a log of the user's recent activities within the application, such as logins, profile updates, and other significant actions.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} GetActivityResponse
// @Router /profile/activity [get]
func _GetActivity() {}

// @Summary Upload User Avatar
// @Description Uploads and sets the authenticated user's avatar. Expects multipart/form-data with the `avatar` file field.
// @Tags User Profile
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} UploadAvatarResponse
// @Router /profile/avatar [post]
func _UploadAvatar() {}
