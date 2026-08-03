package http

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/usecase/dto"
	"sage-backend/internal/users/ports/inbound"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ProfileHandler struct {
	userServ inbound.UserUseCase
}

func NewProfileHandler(userServ inbound.UserUseCase) *ProfileHandler {
	return &ProfileHandler{
		userServ: userServ,
	}
}

// getSessionUser extracts userID and orgID from session
func getSessionUser(c *fiber.Ctx) (userID, orgID string, ok bool) {
	sess, err := config.Store.Get(c)
	if err != nil || sess.Get("userID") == nil {
		return "", "", false
	}
	userID = sess.Get("userID").(string)
	if sess.Get("organizationID") != nil {
		orgID = sess.Get("organizationID").(string)
	}
	return userID, orgID, true
}

// GetProfile returns the comprehensive identity profile
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	profile, err := h.userServ.GetIdentity(c.Context(), userID, orgID)
	if err != nil {
		logger.Error("Error with ProfileHandler.GetProfile: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Profile retrieved successfully", profile)
}

// UpdateProfile updates identity fields (full_name, avatar_url)
func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, _, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.UpdateIdentityRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	var url *string
	fileHeader, _ := c.FormFile("avatar")
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			logger.Error("Error opening avatar file: ", zap.Error(err))
			return response.Error(c, fiber.StatusInternalServerError, "failed to read file", nil)
		}
		defer file.Close()

		mimeType := c.Locals("file_mime_type").(string)
		_, key, err := h.userServ.UploadAvatar(c.Context(), userID, file, mimeType)
		if err != nil {
			logger.Error("Error with ProfileHandler.UploadAvatar: ", zap.Error(err))
			if appErr, ok := err.(*apperrors.ErrorResponse); ok {
				return response.Error(c, appErr.StatusCode, appErr.Message, nil)
			}
			return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
		}
		url = &key
	}

	req.AvatarURL = url

	if err := h.userServ.UpdateIdentity(c.Context(), userID, &req); err != nil {
		logger.Error("Error with ProfileHandler.UpdateProfile: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Profile updated successfully", nil)
}

// GetPreferences returns user preferences
func (h *ProfileHandler) GetPreferences(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	prefs, err := h.userServ.GetPreferences(c.Context(), userID, orgID)
	if err != nil {
		logger.Error("Error with ProfileHandler.GetPreferences: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Preferences retrieved successfully", prefs)
}

// UpdatePreferences updates user preferences
func (h *ProfileHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.userServ.UpdatePreferences(c.Context(), userID, orgID, &req); err != nil {
		logger.Error("Error with ProfileHandler.UpdatePreferences: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Preferences updated successfully", nil)
}

// GetNotifications returns user notification settings
func (h *ProfileHandler) GetNotifications(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	notifs, err := h.userServ.GetNotifications(c.Context(), userID, orgID)
	if err != nil {
		logger.Error("Error with ProfileHandler.GetNotifications: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Notifications retrieved successfully", notifs)
}

// UpdateNotifications updates user notification settings
func (h *ProfileHandler) UpdateNotifications(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.UpdateNotificationsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.userServ.UpdateNotifications(c.Context(), userID, orgID, &req); err != nil {
		logger.Error("Error with ProfileHandler.UpdateNotifications: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Notifications updated successfully", nil)
}

// GetSessions returns all active user sessions
func (h *ProfileHandler) GetSessions(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Get current session ID to mark it
	sess, _ := config.Store.Get(c)
	currentSessionID := ""
	if sess.Get("session_id") != nil {
		currentSessionID = sess.Get("session_id").(string)
	}

	sessions, err := h.userServ.GetSessions(c.Context(), userID, orgID, currentSessionID)
	if err != nil {
		logger.Error("Error with ProfileHandler.GetSessions: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Sessions retrieved successfully", sessions)
}

// RevokeSession revokes a specific session
func (h *ProfileHandler) RevokeSession(c *fiber.Ctx) error {
	userID, _, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	sessionID := c.Params("id")
	if sessionID == "" {
		return response.Error(c, fiber.StatusBadRequest, "session id required", nil)
	}

	// Prevent revoking current session
	sess, _ := config.Store.Get(c)
	if sess.Get("session_id") != nil && sess.Get("session_id").(string) == sessionID {
		return response.Error(c, fiber.StatusForbidden, "cannot revoke current session", nil)
	}

	if err := h.userServ.RevokeSession(c.Context(), sessionID, userID); err != nil {
		logger.Error("Error with ProfileHandler.RevokeSession: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Session revoked successfully", nil)
}

// GetActivity returns user activity/audit logs
func (h *ProfileHandler) GetActivity(c *fiber.Ctx) error {
	userID, orgID, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "25"))

	activity, total, err := h.userServ.GetActivity(c.Context(), userID, orgID, page, pageSize)
	if err != nil {
		logger.Error("Error with ProfileHandler.GetActivity: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Activity retrieved successfully", fiber.Map{
		"data":      activity,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UploadAvatar uploads and sets the user's avatar
func (h *ProfileHandler) UploadAvatar(c *fiber.Ctx) error {
	userID, _, ok := getSessionUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "avatar file required", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.Error("Error opening avatar file: ", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "failed to read file", nil)
	}
	defer file.Close()

	mimeType := c.Locals("file_mime_type").(string)
	url, _, err := h.userServ.UploadAvatar(c.Context(), userID, file, mimeType)
	if err != nil {
		logger.Error("Error with ProfileHandler.UploadAvatar: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Avatar uploaded successfully", fiber.Map{"avatar_url": url})
}
