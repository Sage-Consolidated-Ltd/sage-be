package routes

import (
	"sage-backend/internal/shared/response"
	"sage-backend/internal/users/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, ah *handlers.AuthHandler) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE API SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	RegisterAuthRoutes(v1, ah)

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}

func RegisterAuthRoutes(router fiber.Router, ah *handlers.AuthHandler) {
	auth := router.Group("/auth")

	auth.Post("/register", ah.CreateUser)
	auth.Get("/login/:provider", ah.BeginAuthLogin)
	auth.Get("/callback/:provider", ah.AuthCallback)
}
