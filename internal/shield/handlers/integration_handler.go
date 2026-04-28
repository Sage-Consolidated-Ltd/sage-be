package handlers

import (
	"sage-backend/internal/shield/services"

	"github.com/gofiber/fiber/v2"
)

type IntegrationHandler struct {
	integrationServ services.IntegrationServiceInt
}

func NewIntegrationHandler(integrationServ services.IntegrationServiceInt) *IntegrationHandler {
	return &IntegrationHandler{
		integrationServ: integrationServ,
	}
}

func (h *IntegrationHandler) CreateIntegration(c *fiber.Ctx) error {
	return c.SendString("Create Integration ")
}
