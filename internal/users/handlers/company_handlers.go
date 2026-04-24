package handlers

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
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

// getSessionInfo extracts orgID, userID, and role from session
func getSessionInfo(c *fiber.Ctx) (orgID, userID, role string, ok bool) {
	sess, err := config.Store.Get(c)
	if err != nil {
		return "", "", "", false
	}

	if sess.Get("organizationID") == nil || sess.Get("userID") == nil || sess.Get("role") == nil {
		return "", "", "", false
	}

	orgID = sess.Get("organizationID").(string)
	userID = sess.Get("userID").(string)
	role = sess.Get("role").(string)

	return orgID, userID, role, true
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

	return response.JSON(c, fiber.StatusOK, "invitation accepted successfully", nil)
}

// GetOrganization returns the current organization's core information
// GET /api/v1/organization
func (h *CompanyHandler) GetOrganization(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
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
	orgID, _, role, ok := getSessionInfo(c)
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

	if err := h.companyServ.UpdateOrganization(c.Context(), orgID, role, &req); err != nil {
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
	orgID, _, _, ok := getSessionInfo(c)
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
	orgID, userID, role, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Check permissions
	if !h.companyServ.CanManageMembers(role) {
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
	orgID, _, role, ok := getSessionInfo(c)
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

	if err := h.companyServ.UpdateMemberRole(c.Context(), memberID, orgID, role, req.Role); err != nil {
		logger.Error("Error with CompanyHandler.UpdateMemberRole: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Member role updated successfully", nil)
}

// RemoveMember removes a member from the organization
// DELETE /api/v1/organization/members/:id
func (h *CompanyHandler) RemoveMember(c *fiber.Ctx) error {
	orgID, userID, role, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	memberID := c.Params("id")
	if memberID == "" {
		return response.Error(c, fiber.StatusBadRequest, "member id required", nil)
	}

	if err := h.companyServ.RemoveMember(c.Context(), memberID, orgID, role, userID); err != nil {
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
	orgID, _, _, ok := getSessionInfo(c)
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
	orgID, _, role, ok := getSessionInfo(c)
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

	if err := h.companyServ.UpdateOrganizationSettings(c.Context(), orgID, role, &req); err != nil {
		logger.Error("Error with CompanyHandler.UpdateOrganizationSettings: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Settings updated successfully", nil)
}
