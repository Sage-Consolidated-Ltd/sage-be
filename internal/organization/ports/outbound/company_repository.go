package outbound

import (
	"context"
	"sage-backend/internal/organization/domain"
	"sage-backend/internal/organization/usecase/dto"
	"time"
)

type CompanyRepository interface {
	AddMemberToOrganization(ctx context.Context, organizationId string, userId string, roleId string) error
	GetIndustries(ctx context.Context) (*[]domain.Industry, error)
	GetIndustryByID(ctx context.Context, id string) (*domain.Industry, error)
	InviteMemberToOrganization(ctx context.Context, req *dto.InviteMemberRequest, organizationId string, owner_id string, tokenHash *string, expiresAt time.Time) (*string, error)
	GetByID(ctx context.Context, id string) (*domain.OrganizationInvite, error)
	GetMemberByEmail(ctx context.Context, email string, organizationId string) (*domain.OrganizationMember, error)
	MarkAccepted(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	GetOrganizationRoles(ctx context.Context) (*[]domain.OrganizationRole, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*domain.OrganizationRole, error)
	GetOrganizationByOwnerID(ctx context.Context, ownerId string) (*domain.Organization, error)
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
	ResetMemberMFA(ctx context.Context, memberID, orgID string) error
	UpdateMemberStatus(ctx context.Context, memberID, orgID, status string) error
	GetOrganizationProfile(ctx context.Context, orgID string) (*domain.OrganizationProfileResponse, error)
	UpdateCompanyDetails(ctx context.Context, orgID string, req *dto.UpdateCompanyDetailsRequest) error
	UpdateOrganizationBranding(ctx context.Context, orgID string, req *dto.UpdateOrganizationBrandingRequest) error
	GetValidRoles() []string
	IsValidRole(role string) bool
}
