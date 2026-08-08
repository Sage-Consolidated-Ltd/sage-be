package api

import (
	"context"
	"fmt"

	"sage-backend/internal/app"
	"sage-backend/internal/identity"
	identity_http "sage-backend/internal/identity/adapters/inbound/http"
	"sage-backend/internal/organization"
	org_http "sage-backend/internal/organization/adapters/inbound/http"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	shared_redis "sage-backend/internal/shared/db/redis"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/pkg/jwt"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

// New initializes and constructs the API application with identity and organization modules.
func New() (*app.App, error) {
	cfg := config.SetupAPI()
	config.InitSessionStore(&cfg.BaseConfig)

	database, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisClient, err := shared_redis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Init(&cfg.BaseConfig)

	jwtService := &jwt.JwtService{JwtSecret: cfg.JWTSecret}
	emailClient := mailer.NewEmailClient(&cfg.BaseConfig)

	s3Client, err := s3.NewClient(context.Background(), cfg.S3Bucket, cfg.S3Region)
	var s3Uploader *s3.Uploader
	if err == nil && s3Client != nil {
		s3Uploader = s3.NewUploader(s3Client)
	}

	authMiddleware := &middlewares.AuthMiddleware{}

	orgMod := organization.NewModule(database, nil, emailClient, redisClient, &cfg.BaseConfig, s3Uploader)

	identityMod := identity.NewModule(
		database,
		redisClient,
		cfg,
		jwtService,
		emailClient,
		s3Uploader,
		orgMod.CompanyRepo,
	)

	fiberApp := app.NewFiberApp()

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/users/swagger.json",
		Path:     "/docs/api-docs",
		Title:    "Sage Identity & Organization API Documentation",
		CacheAge: 0,
	}
	fiberApp.Use(swagger.New(swaggerConfig))

	v1 := fiberApp.Group("/api/v1")
	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	identity_http.SetUpRouter(
		v1,
		identityMod.AuthHandler,
		identityMod.ProfileHandler,
		authMiddleware,
	)

	org_http.SetUpRouter(
		v1,
		orgMod.CompanyHandler,
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
