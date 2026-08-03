package app

import (
	"sage-backend/internal/shared/response"
	shield_http "sage-backend/internal/shield/adapters/inbound/http"
	users_http "sage-backend/internal/users/adapters/inbound/http"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

func SetUpRouter(app *App) {
	app.fiber.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE Backend API (Clean Architecture)")
	})

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "SAGE API Documentation",
	}
	app.fiber.Use(swagger.New(swaggerConfig))

	v1 := app.fiber.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	users_http.SetUpRouter(
		v1,
		app.usersModule.AuthHandler,
		app.usersModule.CompanyHandler,
		app.usersModule.ProfileHandler,
		app.authMiddleware,
	)

	shield_http.SetUpRouter(
		v1,
		app.shieldModule.IntegrationHandler,
		app.shieldModule.QualityHandler,
		app.shieldModule.LogsDataHandler,
		app.shieldModule.ParserHandler,
		app.shieldModule.EventHandler,
		app.authMiddleware,
	)

	app.fiber.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}
