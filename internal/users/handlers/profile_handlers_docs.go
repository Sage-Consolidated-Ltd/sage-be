package handlers

import (
	_ "sage-backend/internal/shared/response"
	"sage-backend/internal/users/models"
	_ "sage-backend/internal/users/requests"
)

type UpdateProfileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetActivityResponseData struct {
	Data     []models.UserActivityResponse `json:"data"`
	Total    int                           `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
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
// @Success 200 {object} models.ProfileResponse
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
// @Success 200 {object} models.UserPreferencesResponse
// @Router /profile/preferences [get]
func _GetPreferences() {}

// @Summary Update User Preferences
// @Description Updates the user's preferences for notifications, theme, and other customizable settings.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body requests.UpdatePreferencesRequest true "Update Preferences Request"
// @Success 200 {object} response.Response
// @Router /profile/preferences [patch]
func _UpdatePreferences() {}

// @Summary Get User Notifications
// @Description Retrieves the user's notification settings and history.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.UserNotificationsResponse
// @Router /profile/notifications [get]
func _GetNotifications() {}

// @Summary Update User Notifications
// @Description Updates the user's notification settings.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body requests.UpdateNotificationsRequest true "Update Notifications Request"
// @Success 200 {object} response.Response
// @Router /profile/notifications [patch]
func _UpdateNotifications() {}

// @Summary Get User Session
// @Description Retrieves details about the user's current session, including login time, IP address, and device information.
// @Tags User Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.UserSessionResponse
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
