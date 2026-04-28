package handlers

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"

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

	return response.JSON(c, fiber.StatusOK, "invitation accepted successfully", nil)
}
