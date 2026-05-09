package handlers

import (
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/services"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type IntegrationHandler struct {
	dataSourceServ services.DataSourceServiceInt
}

func NewIntegrationHandler(dataSourceServ services.DataSourceServiceInt) *IntegrationHandler {
	return &IntegrationHandler{
		dataSourceServ: dataSourceServ,
	}
}

func (h *IntegrationHandler) IntegrateDataSource(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}

	var req requests.CreateIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	if err := h.dataSourceServ.CreateDataSource(c.Context(), orgID, req); err != nil {
		logger.Error("Error with IntegrationHandler.IntegrateDataSource: ", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_CREATE_INTEGRATION", err.Error())
	}
	return response.JSON(c, fiber.StatusCreated, "INTEGRATION_CREATED", nil)
}
