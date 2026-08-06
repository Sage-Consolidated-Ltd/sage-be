package inbound

import (
	"context"
	"mime/multipart"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/usecase/dto"
)

type UserUseCase interface {
	// Basic profile (legacy)
	GetProfile(ctx context.Context, userID string) (*dto.GetUserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) error

	// Identity & Access - /api/v1/profile
	GetIdentity(ctx context.Context, userID, orgID string) (*dto.ProfileResponse, error)
	UpdateIdentity(ctx context.Context, userID string, req *dto.UpdateIdentityRequest) error

	// Preferences - /api/v1/profile/preferences
	GetPreferences(ctx context.Context, userID, orgID string) (*domain.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID, orgID string, req *dto.UpdatePreferencesRequest) error

	// Notifications - /api/v1/profile/notifications
	GetNotifications(ctx context.Context, userID, orgID string) (*domain.UserNotifications, error)
	UpdateNotifications(ctx context.Context, userID, orgID string, req *dto.UpdateNotificationsRequest) error

	// Sessions - /api/v1/profile/sessions
	GetSessions(ctx context.Context, userID, orgID, currentSessionID string) ([]*dto.UserSessionResponse, error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	CreateSession(ctx context.Context, userID, orgID, tokenHash, ipAddress, userAgent string) (*domain.UserSession, error)

	// Activity - /api/v1/profile/activity
	GetActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]domain.AuditLog, int, error)
	LogActivity(ctx context.Context, userID, orgID, actionType, resourceType, resourceID string, metadata map[string]interface{}, ipAddress, userAgent string) error

	// Storage
	UploadAvatar(ctx context.Context, userID string, file multipart.File, mimeType string) (string, string, error)
}
