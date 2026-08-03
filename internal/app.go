package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	shared_redis "sage-backend/internal/shared/db/redis"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield"
	"sage-backend/internal/users"
	"sage-backend/pkg/crypto"
	"sage-backend/pkg/jwt"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type App struct {
	config         *config.APIConfig
	fiber          *fiber.App
	usersModule    *users.Module
	shieldModule   *shield.Module
	authMiddleware *middlewares.AuthMiddleware
}

func New() (*App, error) {
	cfg := config.SetupAPI()

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
	encryptor, err := crypto.NewAESEncryptor(cfg.AppEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
	}

	emailClient := mailer.NewEmailClient(&cfg.BaseConfig)

	s3Client, err := s3.NewClient(context.Background(), cfg.S3Bucket, cfg.S3Region)
	var s3Uploader *s3.Uploader
	if err == nil && s3Client != nil {
		s3Uploader = s3.NewUploader(s3Client)
	}

	restyClient := resty.New()
	authMiddleware := &middlewares.AuthMiddleware{}

	usersMod := users.NewModule(
		database,
		redisClient,
		cfg,
		jwtService,
		emailClient,
		s3Uploader,
	)

	shieldMod := shield.NewModule(
		database,
		cfg,
		encryptor,
		restyClient,
	)

	a := &App{
		config: cfg,
		fiber: fiber.New(fiber.Config{
			JSONEncoder: func(v interface{}) ([]byte, error) {
				buf := &bytes.Buffer{}
				encoder := json.NewEncoder(buf)
				encoder.SetEscapeHTML(false)
				err := encoder.Encode(v)
				return bytes.TrimRight(buf.Bytes(), "\n"), err
			},
			EnableTrustedProxyCheck: true,
			TrustedProxies:          []string{"0.0.0.0/0"},
			BodyLimit:               10 * 1024 * 1024,
		}),
		usersModule:    usersMod,
		shieldModule:   shieldMod,
		authMiddleware: authMiddleware,
	}

	a.fiber.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, session_id",
	}))
	a.fiber.Use(recover.New())

	a.fiber.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	SetUpRouter(a)

	return a, nil
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		if err := a.fiber.Listen(fmt.Sprintf(":%s", a.config.PORT)); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	case <-quit:
		logger.Info("Shutting down gracefully...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.fiber.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("Shutdown complete")
	return nil
}
