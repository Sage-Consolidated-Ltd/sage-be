package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/users/handlers"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/routes"
	"sage-backend/internal/users/services"
	"sage-backend/pkg/jwt"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/contrib/swagger"
	_ "sage-backend/docs"
)

// @title           Sage API
// @version         1.0
// @description     Documentation for the Sage API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @host      localhost:3333
// @BasePath  /api/v1
func main() {
	cfg := config.Setup()
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting db: %s", err)
	}
	defer db.Close()

	config.InitSessionStore(cfg)

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0/0"},
		BodyLimit:               5 * 1024 * 1024,
	})
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
        return true
    	},
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowCredentials: true,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, session_id",
	}))
	app.Use(recover.New())

	swaggerConfig := swagger.Config{
		BasePath: "/api/v1",
		FilePath: "./docs/swagger.json",
		Path: "docs",
		Title: "Sage API Documentation",
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

	// repository
	userRepo := repositories.NewUsersRepository(db)

	// services
	authServ := services.NewAuthService(userRepo, jwtService)

	// handlers
	authHandler := handlers.NewAuthHandler(authServ, cfgOAuth, cfg)

	routes.Setup(app, authHandler)

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
