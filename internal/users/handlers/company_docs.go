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
