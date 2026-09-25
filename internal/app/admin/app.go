package admin

import (
	"context"
	"fmt"

	"sage-backend/internal/admin"
	adminHttp "sage-backend/internal/admin/adapters/inbound/http"
	"sage-backend/internal/app"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	sharedRedis "sage-backend/internal/shared/db/redis"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/response"
	"sage-backend/pkg/redoc"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

// New initializes and constructs the Admin platform application.
func New() (*app.App, error) {
	cfg := config.SetupAdmin()

	database, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisClient, err := sharedRedis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Init(&cfg.BaseConfig)

	emailClient := mailer.NewEmailClient(&cfg.BaseConfig)

	adminMod := admin.NewModule(database, redisClient, emailClient, &cfg.BaseConfig)

	if err := adminMod.AuthUseCase.BootstrapSuperAdmin(context.Background(), cfg.SuperAdminEmail, cfg.SuperAdminPassword); err != nil {
		return nil, fmt.Errorf("failed to bootstrap super admin account: %w", err)
	}

	fiberApp := app.NewFiberApp()

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/admin/swagger.json",
		Path:     "/docs/admin-docs",
		Title:    "Sage Admin API Documentation",
		CacheAge: 0,
	}
	fiberApp.Use(swagger.New(swaggerConfig))

	redocConfig := redoc.Config{
		BasePath:   "/api/v1",
		FilePath:   "./docs/admin/swagger.json",
		Path:       "/docs/admin-redoc",
		Title:      "Sage Admin API Documentation",
		SwaggerURL: "/api/v1/docs/admin-docs",
	}
	fiberApp.Use(redoc.New(redocConfig))

	v1 := fiberApp.Group("/api/v1")
	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "Sage Admin API is healthy", nil)
	})

	adminHttp.SetUpAuthRouter(v1, adminMod.AuthHandler, adminMod.AuthMiddleware)

	return &app.App{
		Port:  cfg.PORT,
		Fiber: fiberApp,
	}, nil
}
