package outbound

import (
	"context"
	"sage-backend/internal/identity/domain"
)

type ProfileRepository interface {
	GetUserPreferences(ctx context.Context, userID, orgID string) (*domain.UserPreferences, error)
	UpsertUserPreferences(ctx context.Context, userID, orgID string, prefs *domain.UserPreferences) error
	GetUserNotifications(ctx context.Context, userID, orgID string) (*domain.UserNotifications, error)
	UpsertUserNotifications(ctx context.Context, userID, orgID string, notifs *domain.UserNotifications) error
	GetUserSessions(ctx context.Context, userID, orgID string) ([]domain.UserSession, error)
	GetUserSessionByID(ctx context.Context, sessionID string) (*domain.UserSession, error)
	CreateUserSession(ctx context.Context, session *domain.UserSession) error
	RevokeUserSession(ctx context.Context, sessionID, userID string) error
	UpdateSessionActivity(ctx context.Context, sessionID string) error
	GetUserActivityLog(ctx context.Context, userID, orgID string, page, pageSize int) ([]domain.AuditLog, int, error)
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error
}
