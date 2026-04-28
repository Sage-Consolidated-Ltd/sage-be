package routes

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App,
	ih *handlers.IntegrationHandler,
	m *middlewares.AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE Shield SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}

func RegisterIntegrationRoutes(router fiber.Router, ih *handlers.IntegrationHandler, m *middlewares.AuthMiddleware) {
	integration := router.Group("/integrations")
	integration.Post("/", m.RequireAuth, ih.CreateIntegration)
}
