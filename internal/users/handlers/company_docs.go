package handlers

import (
	_ "sage-backend/internal/shared/response"
	"sage-backend/internal/users/models"
	_ "sage-backend/internal/users/requests"
)

type IndustriesResponse struct {
	Success bool                           `json:"success"`
	Message string                         `json:"message"`
	Data    []models.GetIndustriesResponse `json:"data"`
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
// @Success 200 {object} models.GetOrganizationRolesResponse
// @Router /company/organization-roles [get]
func _GetOrganizationRoles() {}

// @Summary Invite Member to Organization
// @Description Invite a member to the organization by email. Only organization owners can invite members.
// @Tags Company
// @Accept json
// @Produce json
// @Param request body requests.BulkInviteMembersRequest true "Bulk Invite Members Request"
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

type OrganizationMembersResponse struct {
	Success bool                                `json:"success"`
	Message string                              `json:"message"`
	Data    []models.OrganizationMemberResponse `json:"data"`
}

// @Summary List Organization Members
// @Description Retrieve members for the current organization, including role, status, and last active time.
// @Tags Organization
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} OrganizationMembersResponse
// @Router /organization/members [get]
func _ListMembers() {}

// @Summary Update Member Status
// @Description Update a member's status within the organization. Only owners and admins can perform this action.
// @Tags Organization
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Member ID"
// @Param request body requests.UpdateMemberStatusRequest true "Update Member Status Request"
// @Success 200 {object} response.Response
// @Router /organization/members/{id}/status [patch]
func _UpdateMemberStatus() {}

// @Summary Reset Member MFA
// @Description Reset a member's MFA configuration. Only the organization owner can perform this action.
// @Tags Organization
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Member ID"
// @Success 200 {object} response.Response
// @Router /organization/members/{id}/reset-mfa [post]
func _ResetMemberMFA() {}

type PermissionsResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []models.Permission `json:"data"`
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
	Data    []models.PermissionGroup `json:"data"`
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
	Data    []models.CustomRoleResponse `json:"data"`
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
	Data    models.CustomRoleResponse `json:"data"`
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
// @Param request body requests.CreateCustomRoleRequest true "Create Custom Role Request"
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
// @Param request body requests.UpdateCustomRoleRequest true "Update Custom Role Request"
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

// @Summary Switch Active Organization
// @Description Switch the active organization for the current user session. This allows users to change their context between different organizations they belong to.
// @Tags Organization
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} response.Response
// @Router /organization/{id}/switch [post]
func _SwitchActiveOrganization() {}
