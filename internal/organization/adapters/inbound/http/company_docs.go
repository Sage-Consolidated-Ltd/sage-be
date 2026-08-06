package http

import (
	_ "sage-backend/internal/shared/response"
	"sage-backend/internal/organization/domain"
	_ "sage-backend/internal/organization/usecase/dto"
)

type IndustriesResponse struct {
	Success bool                           `json:"success"`
	Message string                         `json:"message"`
	Data    []domain.GetIndustriesResponse `json:"data"`
}

// @Summary Get Industries
// @Description Retrieve a list of industries for company profiles.
// @Tags Company
// @Accept json
// @Produce json
// @Success 200 {object} IndustriesResponse
// @Router /company/industries [get]
func _GetIndustries() {}

// @Summary Get Organization Roles
// @Description Retrieve a list of organization roles for company profiles.
// @Tags Company
// @Accept json
// @Produce json
// @Success 200 {object} domain.GetOrganizationRolesResponse
// @Router /company/organization-roles [get]
func _GetOrganizationRoles() {}

// @Summary Invite Member to Organization
// @Description Invite a member to the organization by email. Only organization owners can invite members.
// @Tags Company
// @Accept json
// @Produce json
// @Param request body dto.BulkInviteMembersRequest true "Bulk Invite Members Request"
// @Success 200 {object} response.Response
// @Router /company/invite [post]
func _InviteMember() {}

// @Summary Accept Organization Invitation
// @Description Accept an invitation to join an organization using the invite ID and token sent via email.
// @Tags Company
// @Accept json
// @Produce json
// @Param invite_id query string true "Invitation ID"
// @Param token query string true "Invitation Token"
// @Success 200 {object} response.Response
// @Router /company/invitations/accept [post]
func _AcceptInvitation() {}

type PermissionsResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []domain.Permission `json:"data"`
}

// @Summary List Permissions
// @Description Retrieve all available permissions in the system.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PermissionsResponse
// @Router /organization/permissions [get]
func ListPermissionsDoc() {}

type PermissionGroupsResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    []domain.PermissionGroup `json:"data"`
}

// @Summary List Permission Groups
// @Description Retrieve all available permission groups in the system.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PermissionGroupsResponse
// @Router /organization/permission-groups [get]
func ListPermissionGroupsDoc() {}

type CustomRolesResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    []domain.CustomRoleResponse `json:"data"`
}

// @Summary List Custom Roles
// @Description Retrieve all custom roles for the current organization.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CustomRolesResponse
// @Router /organization/custom-roles [get]
func ListCustomRolesDoc() {}

type CustomRoleResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    domain.CustomRoleResponse `json:"data"`
}

// @Summary Get Custom Role
// @Description Retrieve a specific custom role by ID.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Custom Role ID"
// @Success 200 {object} CustomRoleResponse
// @Router /organization/custom-roles/{id} [get]
func GetCustomRoleDoc() {}

// @Summary Create Custom Role
// @Description Create a new custom role for the organization. Only owners and admins can create custom roles.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCustomRoleRequest true "Create Custom Role Request"
// @Success 201 {object} CustomRoleResponse
// @Router /organization/custom-roles [post]
func CreateCustomRoleDoc() {}

// @Summary Update Custom Role
// @Description Update an existing custom role. Only owners and admins can update custom roles.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Custom Role ID"
// @Param request body dto.UpdateCustomRoleRequest true "Update Custom Role Request"
// @Success 200 {object} response.Response
// @Router /organization/custom-roles/{id} [patch]
func UpdateCustomRoleDoc() {}

// @Summary Delete Custom Role
// @Description Delete a custom role. Only owners and admins can delete custom roles. System roles cannot be deleted.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Custom Role ID"
// @Success 200 {object} response.Response
// @Router /organization/custom-roles/{id} [delete]
func DeleteCustomRoleDoc() {}

// @Summary Reset Member MFA
// @Description Reset a member's 2FA setup. Only owners and admins can reset member MFA.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Success 200 {object} response.Response
// @Router /organization/members/{id}/reset-mfa [post]
func ResetMemberMFADoc() {}

// @Summary Update Member Status
// @Description Update a member's status (active, suspended, pending). Only owners and admins can update member status.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Member ID"
// @Param request body dto.UpdateMemberStatusRequest true "Update Member Status Request"
// @Success 200 {object} response.Response
// @Router /organization/members/{id}/status [patch]
func UpdateMemberStatusDoc() {}

// @Summary Get Organization Profile
// @Description Retrieve complete organization profile including company details, system environment metadata, and branding/appearance settings.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.OrganizationProfileResponse
// @Router /organization/profile [get]
func GetOrganizationProfileDoc() {}

// @Summary Update Company Details
// @Description Update user-editable company details (organization name, industry, primary contact email, support email). Requires owner or admin role.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateCompanyDetailsRequest true "Update Company Details Request"
// @Success 200 {object} response.Response
// @Router /organization/profile/details [patch]
func UpdateCompanyDetailsDoc() {}

// @Summary Update Organization Branding
// @Description Update logo URLs (light & dark mode) and report display settings. Requires owner or admin role.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateOrganizationBrandingRequest true "Update Organization Branding Request"
// @Success 200 {object} response.Response
// @Router /organization/branding [patch]
func UpdateOrganizationBrandingDoc() {}
