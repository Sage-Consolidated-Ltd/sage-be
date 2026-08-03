package http

import (
	_ "sage-backend/internal/shared/response"
	"sage-backend/internal/users/domain"
	_ "sage-backend/internal/users/usecase/dto"
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
func _ListPermissions() {}

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
func _ListPermissionGroups() {}

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
func _ListCustomRoles() {}

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
func _GetCustomRole() {}

// @Summary Create Custom Role
// @Description Create a new custom role for the organization. Only owners and admins can create custom roles.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCustomRoleRequest true "Create Custom Role Request"
// @Success 201 {object} CustomRoleResponse
// @Router /organization/custom-roles [post]
func _CreateCustomRole() {}

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
func _UpdateCustomRole() {}

// @Summary Delete Custom Role
// @Description Delete a custom role. Only owners and admins can delete custom roles. System roles cannot be deleted.
// @Tags Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Custom Role ID"
// @Success 200 {object} response.Response
// @Router /organization/custom-roles/{id} [delete]
func _DeleteCustomRole() {}
