package handlers

import (
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
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