package handlers

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type CompanyHandler struct {
	companyServ services.CompanyServicesInt
}

func NewCompanyHandler(companyServ services.CompanyServicesInt) *CompanyHandler {
	return &CompanyHandler{
		companyServ: companyServ,
	}
}

func (h *CompanyHandler) GetIndustries(c *fiber.Ctx) error {
	industries, err := h.companyServ.GetIndustries()
	if err != nil {
		logger.Error("Error with CompanyHandler.GetIndustries: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}
	return response.JSON(c, fiber.StatusOK, "Industries retrieved successfully", industries)
}
func (h *CompanyHandler) GetOrganizationRoles(c *fiber.Ctx) error {
	roles, err := h.companyServ.GetOrganizationRoles()
	if err != nil {
		logger.Error("Error with CompanyHandler.GetOrganizationRoles: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}
	return response.JSON(c, fiber.StatusOK, "Organization roles retrieved successfully", roles)
}
func (h *CompanyHandler) InviteMember(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	ownerId, ok := sess.Get("userID").(string)
	if !ok || ownerId == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req requests.BulkInviteMembersRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.InviteUserToOrganization(c.Context(), &req, ownerId); err != nil {
		logger.Error("Error with CompanyHandler:InviteMember, ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.JSON(c, fiber.StatusOK, "invitations sent successfully", nil)
}

func (h *CompanyHandler) AcceptInvitation(c *fiber.Ctx) error {
	inviteId := c.Query("invite_id")
	token := c.Query("token")

	if inviteId == "" || token == "" {
		return response.Error(c, fiber.StatusBadRequest, "invite_id and token are required", nil)
	}

	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	userId, ok := sess.Get("userID").(string)
	if !ok || userId == "" {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if err := h.companyServ.AcceptInvitation(c.Context(), inviteId, token, userId); err != nil {
		logger.Error("Error with CompanyHandler.AcceptInvitation: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Delete("pending_email_verification")
	sess.Save()

	return response.JSON(c, fiber.StatusOK, "invitation accepted successfully", nil)
}

// GetOrganization returns the current organization's core information
// GET /api/v1/organization
func (h *CompanyHandler) GetOrganization(c *fiber.Ctx) error {
	orgID, _, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	org, err := h.companyServ.GetOrganization(c.Context(), orgID)
	if err != nil {
		logger.Error("Error with CompanyHandler.GetOrganization: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Organization retrieved successfully", org)
}

// UpdateOrganization updates organization metadata
// PATCH /api/v1/organization
func (h *CompanyHandler) UpdateOrganization(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req requests.UpdateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.UpdateOrganization(c.Context(), orgID, roleInOrg, &req); err != nil {
		logger.Error("Error with CompanyHandler.UpdateOrganization: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Organization updated successfully", nil)
}

// ListMembers returns all members in the organization
// GET /api/v1/organization/members
func (h *CompanyHandler) ListMembers(c *fiber.Ctx) error {
	orgID, _, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Parse query params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "25"))
	role := c.Query("role", "")
	search := c.Query("search", "")

	req := requests.ListMembersRequest{
		Page:     page,
		PageSize: pageSize,
		Role:     role,
		Search:   search,
	}

	members, total, err := h.companyServ.ListMembers(c.Context(), orgID, &req)
	if err != nil {
		logger.Error("Error with CompanyHandler.ListMembers: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Members retrieved successfully", fiber.Map{
		"data":      members,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// InviteMembers invites users to join the organization
// POST /api/v1/organization/members/invite
func (h *CompanyHandler) InviteMembers(c *fiber.Ctx) error {
	orgID, roleInOrg, userID, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Check permissions
	if !h.companyServ.CanManageMembers(roleInOrg) {
		return response.Error(c, fiber.StatusForbidden, "only owner or admin can invite members", nil)
	}

	var req requests.InviteMembersRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	result, err := h.companyServ.InviteMembers(c.Context(), orgID, userID, &req)
	if err != nil {
		logger.Error("Error with CompanyHandler.InviteMembers: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Invitations processed", result)
}

// UpdateMemberRole updates a member's role
// PATCH /api/v1/organization/members/:id/role
func (h *CompanyHandler) UpdateMemberRole(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	memberID := c.Params("id")
	if memberID == "" {
		return response.Error(c, fiber.StatusBadRequest, "member id required", nil)
	}

	var req requests.UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.UpdateMemberRole(c.Context(), memberID, orgID, roleInOrg, req.Role); err != nil {
		logger.Error("Error with CompanyHandler.UpdateMemberRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Member role updated successfully", nil)
}

// ResetMemberMFA resets a member's MFA configuration.
// POST /api/v1/organization/members/:id/reset-mfa
func (h *CompanyHandler) ResetMemberMFA(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	memberID := c.Params("id")
	if memberID == "" {
		return response.Error(c, fiber.StatusBadRequest, "member id required", nil)
	}

	if err := h.companyServ.ResetMemberMFA(c.Context(), memberID, orgID, roleInOrg); err != nil {
		logger.Error("Error with CompanyHandler.ResetMemberMFA: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Member MFA reset successfully", nil)
}

// RemoveMember removes a member from the organization
// DELETE /api/v1/organization/members/:id
func (h *CompanyHandler) RemoveMember(c *fiber.Ctx) error {
	orgID, roleInOrg, userID, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	memberID := c.Params("id")
	if memberID == "" {
		return response.Error(c, fiber.StatusBadRequest, "member id required", nil)
	}

	if err := h.companyServ.RemoveMember(c.Context(), memberID, orgID, roleInOrg, userID); err != nil {
		logger.Error("Error with CompanyHandler.RemoveMember: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Member removed successfully", nil)
}

// GetOrganizationSettings returns organization-wide settings
// GET /api/v1/organization/settings
func (h *CompanyHandler) GetOrganizationSettings(c *fiber.Ctx) error {
	orgID, _, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	settings, err := h.companyServ.GetOrganizationSettings(c.Context(), orgID)
	if err != nil {
		logger.Error("Error with CompanyHandler.GetOrganizationSettings: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Settings retrieved successfully", settings)
}

// UpdateOrganizationSettings updates organization settings
// PATCH /api/v1/organization/settings
func (h *CompanyHandler) UpdateOrganizationSettings(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req requests.UpdateOrganizationSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.UpdateOrganizationSettings(c.Context(), orgID, roleInOrg, &req); err != nil {
		logger.Error("Error with CompanyHandler.UpdateOrganizationSettings: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Settings updated successfully", nil)
}

// ListPermissions returns all available permissions
// GET /api/v1/organization/permissions
func (h *CompanyHandler) ListPermissions(c *fiber.Ctx) error {
	permissions, err := h.companyServ.ListPermissions(c.Context())
	if err != nil {
		logger.Error("Error with CompanyHandler.ListPermissions: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Permissions retrieved successfully", permissions)
}

// ListPermissionGroups returns all permission groups
// GET /api/v1/organization/permission-groups
func (h *CompanyHandler) ListPermissionGroups(c *fiber.Ctx) error {
	groups, err := h.companyServ.ListPermissionGroups(c.Context())
	if err != nil {
		logger.Error("Error with CompanyHandler.ListPermissionGroups: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Permission groups retrieved successfully", groups)
}

// ListCustomRoles returns all custom roles for the organization
// GET /api/v1/organization/custom-roles
func (h *CompanyHandler) ListCustomRoles(c *fiber.Ctx) error {
	orgID, _, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	roles, err := h.companyServ.ListCustomRoles(c.Context(), orgID)
	if err != nil {
		logger.Error("Error with CompanyHandler.ListCustomRoles: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Custom roles retrieved successfully", roles)
}

// GetCustomRole returns a specific custom role
// GET /api/v1/organization/custom-roles/:id
func (h *CompanyHandler) GetCustomRole(c *fiber.Ctx) error {
	orgID, _, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	roleID := c.Params("id")
	if roleID == "" {
		return response.Error(c, fiber.StatusBadRequest, "role id is required", nil)
	}

	role, err := h.companyServ.GetCustomRole(c.Context(), roleID, orgID)
	if err != nil {
		logger.Error("Error with CompanyHandler.GetCustomRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Custom role retrieved successfully", role)
}

// CreateCustomRole creates a new custom role
// POST /api/v1/organization/custom-roles
func (h *CompanyHandler) CreateCustomRole(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if !h.companyServ.CanManageMembers(roleInOrg) {
		return response.Error(c, fiber.StatusForbidden, "only owner or admin can create custom roles", nil)
	}

	var req requests.CreateCustomRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	roleResp, err := h.companyServ.CreateCustomRole(c.Context(), orgID, &req)
	if err != nil {
		logger.Error("Error with CompanyHandler.CreateCustomRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusCreated, "Custom role created successfully", roleResp)
}

// UpdateCustomRole updates a custom role
// PATCH /api/v1/organization/custom-roles/:id
func (h *CompanyHandler) UpdateCustomRole(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if !h.companyServ.CanManageMembers(roleInOrg) {
		return response.Error(c, fiber.StatusForbidden, "only owner or admin can update custom roles", nil)
	}

	roleID := c.Params("id")
	if roleID == "" {
		return response.Error(c, fiber.StatusBadRequest, "role id is required", nil)
	}

	var req requests.UpdateCustomRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.UpdateCustomRole(c.Context(), roleID, orgID, &req); err != nil {
		logger.Error("Error with CompanyHandler.UpdateCustomRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Custom role updated successfully", nil)
}

// DeleteCustomRole deletes a custom role
// DELETE /api/v1/organization/custom-roles/:id
func (h *CompanyHandler) DeleteCustomRole(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if !h.companyServ.CanManageMembers(roleInOrg) {
		return response.Error(c, fiber.StatusForbidden, "only owner or admin can delete custom roles", nil)
	}

	roleID := c.Params("id")
	if roleID == "" {
		return response.Error(c, fiber.StatusBadRequest, "role id is required", nil)
	}

	if err := h.companyServ.DeleteCustomRole(c.Context(), roleID, orgID); err != nil {
		logger.Error("Error with CompanyHandler.DeleteCustomRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Custom role deleted successfully", nil)
}

func (h *CompanyHandler) UpdateMemberStatus(c *fiber.Ctx) error {
	orgID, roleInOrg, _, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if !h.companyServ.CanManageMembers(roleInOrg) {
		return response.Error(c, fiber.StatusForbidden, "only owner or admin can update member status", nil)
	}

	memberID := c.Params("id")
	if memberID == "" {
		return response.Error(c, fiber.StatusBadRequest, "member id required", nil)
	}

	var req requests.UpdateMemberStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.companyServ.UpdateMemberStatus(c.Context(), memberID, orgID, roleInOrg, req.Status); err != nil {
		logger.Error("Error with CompanyHandler.UpdateMemberStatus: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Member status updated successfully", nil)
}

func (h *CompanyHandler) SwitchActiveOrganization(c *fiber.Ctx) error {
	_, _, userID, _, ok := middlewares.GetSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	organizationID := c.Params("id")
	if organizationID == "" {
		return response.Error(c, fiber.StatusBadRequest, "organization id required", nil)
	}

	if _, err := h.companyServ.GetOrganizationByID(c.Context(), organizationID); err != nil {
		logger.Error("Could not fetch organization at CompanyHandler.SwitchActiveOrganization", zap.Error(err))
		return response.Error(c, fiber.StatusNotFound, "organization not found", nil)
	}

	isMember, err := h.companyServ.CheckMemberInOrganization(c.Context(), userID, organizationID)
	if err != nil {
		logger.Error("Could not check member in organization at CompanyHandler.SwitchActiveOrganization", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	if !isMember {
		return response.Error(c, fiber.StatusForbidden, "user is not a member of the organization", nil)
	}

	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	sess.Set("activeOrgID", organizationID)
	if err := sess.Save(); err != nil {
		logger.Error("Could not save session at CompanyHandler.SwitchActiveOrganization", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Switched active organization successfully", fiber.Map{
		"organization_id": organizationID,
	})
}	
