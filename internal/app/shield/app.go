package shield

import (
	"fmt"

	"sage-backend/internal/app"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	shared_redis "sage-backend/internal/shared/db/redis"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield"
	shield_http "sage-backend/internal/shield/adapters/inbound/http"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

// New initializes and constructs the Shield API application.
func New() (*app.App, error) {
	cfg := config.SetupShield()
	config.InitSessionStore(&cfg.BaseConfig)

	database, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	_, err = shared_redis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Init(&cfg.BaseConfig)

	encryptor, err := crypto.NewAESEncryptor(cfg.AppEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
	}

	restyClient := resty.New()
	authMiddleware := &middlewares.AuthMiddleware{}

	shieldMod := shield.NewModule(
		database,
		&config.APIConfig{BaseConfig: cfg.BaseConfig},
		encryptor,
		restyClient,
	)

	fiberApp := app.NewFiberApp()

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/shield/swagger.json",
		Path:     "docs/shield-docs",
		Title:    "Sage Shield API Documentation",
		CacheAge: 0,
	}
	fiberApp.Use(swagger.New(swaggerConfig))

	v1 := fiberApp.Group("/api/v1")
	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "Shield API is Healthy", nil)
	})

	shield_http.SetUpRouter(
		v1,
		shieldMod.IntegrationHandler,
		shieldMod.QualityHandler,
		shieldMod.LogsDataHandler,
		shieldMod.ParserHandler,
		shieldMod.EventHandler,
		authMiddleware,
	)

	fiberApp.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})

	return &app.App{
		Port:  cfg.PORT,
		Fiber: fiberApp,
	}, nil
}
