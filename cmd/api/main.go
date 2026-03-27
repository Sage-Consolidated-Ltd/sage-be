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
)

func main() {
	cfg := config.Setup()
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting db: %s", err)
	}
	defer db.Close()

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0/0"},
		BodyLimit:               5 * 1024 * 1024,
	})
	app.Use(cors.New(cors.Config{
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, session_id",
	}))
	app.Use(recover.New())

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	jwtService := &jwt.JwtService{
		JwtSecret: cfg.JWTSecret,
	}

	// repository
	userRepo := repositories.NewUsersRepository(db)

	// services
	authServ := services.NewAuthService(userRepo, jwtService)

	// handlers
	authHandler := handlers.NewAuthHandler(authServ)

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
