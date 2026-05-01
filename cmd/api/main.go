package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/db/redis"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/users/handlers"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/routes"
	"sage-backend/internal/users/services"
	"sage-backend/pkg/jwt"
	"strings"
	"syscall"

	_ "sage-backend/docs"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title           Sage API
// @version         1.0
// @description     Documentation for the Sage API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @host      backend.sageconsolidated.com
// @BasePath  /api/v1
func main() {
	cfg := config.SetupAPI()
	db, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Error connecting db: %s", err)
	}
	defer db.Close()

	redis, err := redis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Error connecting redis: %s", err)
	}

	config.InitSessionStore(&cfg.BaseConfig)

	logger.Init(&cfg.BaseConfig)

	mailer := mailer.NewEmailClient(&cfg.BaseConfig)

	app := fiber.New(fiber.Config{
		JSONEncoder: func(v interface{}) ([]byte, error) {
			buf := &bytes.Buffer{}
			encoder := json.NewEncoder(buf)
			encoder.SetEscapeHTML(false)
			err := encoder.Encode(v)
			return bytes.TrimRight(buf.Bytes(), "\n"), err
		},
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0/0"},
		BodyLimit:               5 * 1024 * 1024,
	})
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, session_id",
	}))
	app.Use(recover.New())

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/swagger.json",
		Path:     "docs/api-docs",
		Title:    "Sage API Documentation",
	}

	app.Use(swagger.New(swaggerConfig))

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	jwtService := &jwt.JwtService{
		JwtSecret: cfg.JWTSecret,
	}

	cfgOAuth := config.NewOAuthConfig(
		cfg.GoogleClientId,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectUrl,
		cfg.GitHubClientId,
		cfg.GitHubClientSecret,
		cfg.GitHubRedirectUrl,
		cfg.AzureClientId,
		cfg.AzureClientSecret,
		cfg.AzureRedirectUrl,
	)

	// middlewares
	authMiddleware := &middlewares.AuthMiddleware{}

	// repository
	userRepo := repositories.NewUsersRepository(db)
	companyRepo := repositories.NewCompanyRepository(db)

	// services
	authServ := services.NewAuthService(userRepo, jwtService, cfg, redis, companyRepo, mailer)
	companyServ := services.NewCompanyServices(companyRepo, userRepo, mailer, redis, &cfg.BaseConfig)
	profileRepo := repositories.NewProfileRepository(db)
	userServ := services.NewUsersServices(userRepo, profileRepo, redis)

	// handlers
	authHandler := handlers.NewAuthHandler(authServ, cfgOAuth, cfg)
	companyHandler := handlers.NewCompanyHandler(companyServ)
	profileHandler := handlers.NewProfileHandler(userServ)

	routes.Setup(app, authHandler, companyHandler, profileHandler, authMiddleware)

	port := strings.TrimSpace(cfg.PORT)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	_ = <-c
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()
}
