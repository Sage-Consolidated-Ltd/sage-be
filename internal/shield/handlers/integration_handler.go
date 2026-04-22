package handlers

import "github.com/gofiber/fiber/v2"

type IntegrationHandler struct{}

func NewIntegrationHandler() *IntegrationHandler{
	return &IntegrationHandler{}
}

func (h *IntegrationHandler) CreateIntegration(c *fiber.Ctx) error {
	return c.SendString("Create Integration")
}