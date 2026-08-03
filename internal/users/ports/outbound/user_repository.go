package outbound

import (
	"context"
	"mime/multipart"
	"sage-backend/internal/users/domain"
	"sage-backend/internal/users/usecase/dto"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest, hash string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	MarkEmailVerified(ctx context.Context, email string) error
	CreateUserWithOrganization(ctx context.Context, req *dto.CreateUserRequest, hash string) error
	GetUserOrganizations(ctx context.Context, userId string) (*[]domain.Organization, error)
	Enable2FA(ctx context.Context, secret string, userID string) error
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	UpdateUserPassword(ctx context.Context, email string, hash string) error
	OnboardUserWithTransaction(ctx context.Context, req *dto.OnboardingRequest, hash string) (*domain.User, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateProfileRequest) error
}

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
	GetUserActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]domain.AuditLog, int, error)
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error
}

type CompanyRepository interface {
	GetIndustries(ctx context.Context) (*[]domain.Industry, error)
	GetOrganizationRoles(ctx context.Context) (*[]domain.OrganizationRole, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*domain.OrganizationRole, error)
	GetIndustryByID(ctx context.Context, id string) (*domain.Industry, error)
	InviteMemberToOrganization(ctx context.Context, req *dto.InviteMemberRequest, organizationId string, owner_id string, tokenHash *string, expiresAt time.Time) (*string, error)
	GetByID(ctx context.Context, id string) (*domain.OrganizationInvite, error)
	MarkAccepted(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	GetMemberByEmail(ctx context.Context, email string, organizationId string) (*domain.OrganizationMember, error)
	GetOrganizationByOwnerID(ctx context.Context, ownerId string) (*domain.Organization, error)
	AddMemberToOrganization(ctx context.Context, organizationId string, userId string, roleId string) error
	GetOrganizationByID(ctx context.Context, orgID string) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, orgID string, req *dto.UpdateOrganizationRequest) error
	ListMembers(ctx context.Context, orgID string, page, pageSize int, role, search string) ([]domain.OrganizationMember, int, error)
	GetMemberByID(ctx context.Context, memberID, orgID string) (*domain.OrganizationMember, error)
	AddMember(ctx context.Context, orgID, userID, role, invitedBy string) (*domain.OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, memberID, orgID, newRole string) error
	RemoveMember(ctx context.Context, memberID, orgID string) error
	GetMemberCountByRole(ctx context.Context, orgID, role string) (int, error)
	GetOrganizationSettings(ctx context.Context, orgID string) (*domain.OrganizationSettings, error)
	CreateOrganizationSettings(ctx context.Context, orgID string) (*domain.OrganizationSettings, error)
	UpdateOrganizationSettings(ctx context.Context, orgID string, req *dto.UpdateOrganizationSettingsRequest) error
	GetValidRoles() []string
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	ListPermissionGroups(ctx context.Context) ([]domain.PermissionGroup, error)
	ListCustomRoles(ctx context.Context, orgID string) ([]domain.CustomRole, error)
	GetCustomRoleByID(ctx context.Context, roleID, orgID string) (*domain.CustomRole, error)
	CreateCustomRole(ctx context.Context, orgID string, req *dto.CreateCustomRoleRequest) (*domain.CustomRole, error)
	UpdateCustomRole(ctx context.Context, roleID, orgID string, req *dto.UpdateCustomRoleRequest) error
	DeleteCustomRole(ctx context.Context, roleID, orgID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)
	GetRolePermissionGroups(ctx context.Context, roleID string) ([]string, error)
	SetRolePermissionGroups(ctx context.Context, roleID string, groupIDs []string) error
	SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	IsValidRole(role string) bool
	CanManageMembers(actorRole string) bool
	CanUpdateSettings(actorRole string) bool
}

type Mailer interface {
	SendVerificationEmail(to []string, data interface{}) error
}

type StorageUploader interface {
	UploadAvatar(ctx context.Context, file multipart.File, userID, mimeType string) (string, string, error)
	GenerateSignedURL(ctx context.Context, key string) (string, error)
}
