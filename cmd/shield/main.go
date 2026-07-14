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
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/handlers"
	shieldMiddlewares "sage-backend/internal/shield/middlewares"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/routes"
	"sage-backend/internal/shield/scheduler"
	"sage-backend/internal/shield/services"
	"sage-backend/internal/shield/tasks"
	"strings"
	"syscall"

	_ "sage-backend/docs/shield"
	"sage-backend/pkg/crypto"

	"context"

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

// @host      backend.sageconsolidated.com
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

	redisRef := os.Getenv("REDIS_DB_URL")
	if redisRef == "" {
		redisRef = os.Getenv("REDIS_ADDR")
	}
	if redisRef == "" {
		redisRef = "localhost:6379"
	}
	taskClient := tasks.NewTaskClient(redisRef)
	defer taskClient.Close()

	redisClient, err := redis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Error connecting redis: %s", err)
	}

	logger.Init(&cfg.BaseConfig)

	_ = mailer.NewEmailClient(&cfg.BaseConfig)

	s3Client, _ := s3.NewClient(s3.S3Config{
		Region:          cfg.BaseConfig.S3Region,
		Bucket:          cfg.BaseConfig.S3Bucket,
		AccessKeyID:     cfg.BaseConfig.S3AccessKey,
		SecretAccessKey: cfg.BaseConfig.S3SecretKey,
		PresignExpiry:   24 * 60,
		MaxFileSizeMB:   20 * 1024 * 1024,
	})
	uploader := s3.NewUploader(s3Client)

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
	uploadLogRepo := repositories.NewLogUploadRepository(db)
	parsedLogRepo := repositories.NewParsedLogRepository(db)

	dataQualityServ := services.NewDataQualityService(dataQualtyRepo, parserRepo, dataSourceRepo, ingestionRepo)
	logsDataServ := services.NewLogsDataService(dataSourceRepo, eventRepo, ingestionRepo, taskClient)
	logsServ := services.NewLogsService(eventRepo, dataSourceRepo, ingestionRepo)
	parserServ := services.NewParserService(parserRepo, eventRepo, dataSourceRepo, ingestionRepo)
	integrationServ := services.NewDataSourceService(dataSourceRepo, integrationRepo, encryptor, restyClient)
	logUploadServ := services.NewUploadService(uploader, uploadLogRepo, taskClient, dataSourceRepo)
	parsedLogServ := services.NewParsedLogService(parsedLogRepo, parsedLogRepo)

	integrationHandler := handlers.NewIntegrationHandler(integrationServ)
	eventHandler := handlers.NewEventHandler(logsServ)
	logsDataHandler := handlers.NewLogsDataHandlerWithService(logsDataServ)
	parserHandler := handlers.NewParserHandler(parserServ)
	qualityHandler := handlers.NewQualityHandler(dataQualityServ)
	uploadLogHandler := handlers.NewUploadHandler(logUploadServ)
	parsedLogHandler := handlers.NewParsedLogHandler(*parsedLogServ)

	// Initialize provider scheduler for periodic syncs (300 seconds = 5 minutes)
	providerScheduler := scheduler.NewProviderScheduler(taskClient, dataSourceRepo, 300)

	routes.Setup(
		app,
		integrationHandler,
		qualityHandler,
		logsDataHandler,
		parserHandler,
		eventHandler,
		uploadLogHandler,
		parsedLogHandler,
		shieldMiddlewares.NewRateLimiter(
			3,
			logger.Default(),
			shieldMiddlewares.NewRedisLimiterStore(redisClient),
		),
		authMiddleware,
	)

	port := strings.TrimSpace(cfg.PORT)

	// Start provider scheduler in background
	providerScheduler.Start(context.Background())

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	_ = <-c
	fmt.Println("Gracefully shutting down...")
	providerScheduler.Stop()
	_ = app.Shutdown()
}
