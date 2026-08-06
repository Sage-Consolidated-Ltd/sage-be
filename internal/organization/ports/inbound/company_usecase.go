package inbound

import (
	"context"
	"mime/multipart"
	"sage-backend/internal/organization/domain"
	"sage-backend/internal/organization/usecase/dto"
)

type CompanyUseCase interface {
	GetIndustries(ctx context.Context) (*[]domain.GetIndustriesResponse, error)
	GetOrganizationRoles(ctx context.Context) (*[]domain.GetOrganizationRolesResponse, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*domain.GetOrganizationRolesResponse, error)
	InviteUserToOrganization(ctx context.Context, req *dto.BulkInviteMembersRequest, ownerId string) error
	AcceptInvitation(ctx context.Context, inviteId string, token string, userId string) error
	GetOrganization(ctx context.Context, orgID string) (*domain.GetOrganizationResponse, error)
	GetOrganizationByID(ctx context.Context, orgID string) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationRequest) error
	ListMembers(ctx context.Context, orgID string, req *dto.ListMembersRequest) ([]*domain.OrganizationMemberResponse, int, error)
	GetMemberByID(ctx context.Context, memberID, orgID string) (*domain.OrganizationMember, error)
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
	ResetMemberMFA(ctx context.Context, memberID, orgID, actorRole string) error
	UpdateMemberStatus(ctx context.Context, memberID, orgID, actorRole, status string) error
	GetOrganizationProfile(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error)
	UpdateCompanyDetails(ctx context.Context, orgID, actorRole string, req *dto.UpdateCompanyDetailsRequest) error
	UpdateOrganizationBranding(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationBrandingRequest) error
	UploadBrandingLogo(ctx context.Context, orgID, actorRole, logoType string, file multipart.File, mimeType string) (string, error)
}

