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
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shield/handlers"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/routes"
	"sage-backend/internal/shield/services"
	"strings"
	"syscall"

	_ "sage-backend/docs/shield"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title           Sage API (Shield)
// @version         1.0
// @description     Documentation for the Sage API (Shield).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session_id

// @host      localhost:3335
// @BasePath  /api/v1
func main() {
	cfg := config.SetupShield()
	db, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err)
	}
	defer db.Close()

	config.InitSessionStore(&cfg.BaseConfig)

	restyClient := resty.New()

	logger.Init(&cfg.BaseConfig)

	_ = mailer.NewEmailClient(&cfg.BaseConfig)

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
		FilePath: "./docs/shield/swagger.json",
		Path:     "docs/shield-docs",
		Title:    "Sage API Documentation",
	}

	app.Use(swagger.New(swaggerConfig))

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	encryptor, err := crypto.NewAESEncryptor(cfg.AppEncryptionKey)
	if err != nil {
		log.Fatalf("Error initializing encryptor: %s", err)
	}

	authMiddleware := &middlewares.AuthMiddleware{}

	integrationRepo := repositories.NewIntegrationRepository(db)
	dataQualtyRepo := repositories.NewDataQualityRepository(db)
	dataSourceRepo := repositories.NewDataSourceRepository(db)
	eventRepo := repositories.NewSecurityEventRepository(db)
	ingestionRepo := repositories.NewIngestionJobRepository(db)
	parserRepo := repositories.NewParserRepository(db)

	dataQualityServ := services.NewDataQualityService(dataQualtyRepo, parserRepo, dataSourceRepo, ingestionRepo)
	logsDataServ := services.NewLogsDataService(dataSourceRepo, eventRepo, ingestionRepo)
	logsServ := services.NewLogsService(eventRepo, dataSourceRepo, ingestionRepo)
	parserServ := services.NewParserService(parserRepo, eventRepo, dataSourceRepo, ingestionRepo)
	integrationServ := services.NewDataSourceService(dataSourceRepo, integrationRepo, encryptor, restyClient)

	integrationHandler := handlers.NewIntegrationHandler(integrationServ)
	eventHandler := handlers.NewEventHandler(logsServ)
	logsDataHandler := handlers.NewLogsDataHandlerWithService(logsDataServ)
	parserHandler := handlers.NewParserHandler(parserServ)
	qualityHandler := handlers.NewQualityHandler(dataQualityServ)

	routes.Setup(app, integrationHandler, qualityHandler, logsDataHandler, parserHandler, eventHandler, authMiddleware)

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
