package inbound

import (
	"context"
	"mime/multipart"
	"sage-backend/internal/users/domain"
	"sage-backend/internal/users/usecase/dto"
)

type AuthUseCase interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) error
	CreateUserWithOrganization(ctx context.Context, req *dto.OnboardingRequest) (*domain.GetUserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*domain.GetUserResponse, error)
	OAuthLogin(ctx context.Context, payload *dto.CreateUserRequest) (*domain.GetUserResponse, error)
	FetchExternalUser(ctx context.Context, provider, accessToken string) (*domain.ExternalUser, error)
	Generate2FA(ctx context.Context, email string) (string, string, error)
	Enabled2FA(ctx context.Context, code string, secret string, userID string) error
	Verify2FA(ctx context.Context, code, userID string) error
	ForgotPassword(ctx context.Context, email string) error
	VerifyResetToken(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest, token string) error
	SendEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
}

type UserUseCase interface {
	GetProfile(ctx context.Context, userID string) (*domain.GetUserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) error
	GetIdentity(ctx context.Context, userID, orgID string) (*domain.ProfileResponse, error)
	UpdateIdentity(ctx context.Context, userID string, req *dto.UpdateIdentityRequest) error
	GetPreferences(ctx context.Context, userID, orgID string) (*domain.UserPreferencesResponse, error)
	UpdatePreferences(ctx context.Context, userID, orgID string, req *dto.UpdatePreferencesRequest) error
	GetNotifications(ctx context.Context, userID, orgID string) (*domain.UserNotificationsResponse, error)
	UpdateNotifications(ctx context.Context, userID, orgID string, req *dto.UpdateNotificationsRequest) error
	GetSessions(ctx context.Context, userID, orgID, currentSessionID string) ([]*domain.UserSessionResponse, error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	CreateSession(ctx context.Context, userID, orgID, tokenHash, ipAddress, userAgent string) (*domain.UserSession, error)
	GetActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]*domain.UserActivityResponse, int, error)
	LogActivity(ctx context.Context, userID, orgID, actionType, resourceType, resourceID string, metadata map[string]interface{}, ipAddress, userAgent string) error
	UploadAvatar(ctx context.Context, userID string, file multipart.File, mimeType string) (string, string, error)
}

type CompanyUseCase interface {
	GetIndustries(ctx context.Context) (*[]domain.GetIndustriesResponse, error)
	GetOrganizationRoles(ctx context.Context) (*[]domain.GetOrganizationRolesResponse, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*domain.GetOrganizationRolesResponse, error)
	InviteUserToOrganization(ctx context.Context, req *dto.BulkInviteMembersRequest, ownerId string) error
	AcceptInvitation(ctx context.Context, inviteId string, token string, userId string) error
	GetOrganization(ctx context.Context, orgID string) (*domain.GetOrganizationResponse, error)
	UpdateOrganization(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationRequest) error
	ListMembers(ctx context.Context, orgID string, req *dto.ListMembersRequest) ([]*domain.OrganizationMemberResponse, int, error)
	InviteMembers(ctx context.Context, orgID, invitedBy string, req *dto.InviteMembersRequest) (*dto.InviteMembersResult, error)
	UpdateMemberRole(ctx context.Context, memberID, orgID, actorRole, newRole string) error
	RemoveMember(ctx context.Context, memberID, orgID, actorRole, actorUserID string) error
	GetOrganizationSettings(ctx context.Context, orgID string) (*domain.OrganizationSettingsResponse, error)
	UpdateOrganizationSettings(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationSettingsRequest) error
	CanManageMembers(role string) bool
	CanUpdateSettings(role string) bool
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	ListPermissionGroups(ctx context.Context) ([]domain.PermissionGroup, error)
	ListCustomRoles(ctx context.Context, orgID string) ([]domain.CustomRoleResponse, error)
	GetCustomRole(ctx context.Context, roleID, orgID string) (*domain.CustomRoleResponse, error)
	CreateCustomRole(ctx context.Context, orgID string, req *dto.CreateCustomRoleRequest) (*domain.CustomRoleResponse, error)
	UpdateCustomRole(ctx context.Context, roleID, orgID string, req *dto.UpdateCustomRoleRequest) error
	DeleteCustomRole(ctx context.Context, roleID, orgID string) error
}
