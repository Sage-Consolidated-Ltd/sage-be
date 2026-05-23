package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/tasks"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.SetupShield()

	// Initialize database
	db, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	serverOpt := asynq.RedisClientOpt{Addr: redisRef}
	if strings.Contains(redisRef, "://") {
		if parsed, err := redis.ParseURL(redisRef); err == nil {
			serverOpt = asynq.RedisClientOpt{
				Addr:     parsed.Addr,
				Password: parsed.Password,
				DB:       parsed.DB,
			}
		}
	}

	encryptor, err := crypto.NewAESEncryptor(cfg.AppEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryptor: %v", err)
	}
	// Initialize repositories
	jobRepo := repositories.NewIngestionJobRepository(db)
	integrationRepo := repositories.NewIntegrationRepository(db)
	dataSourceRepo := repositories.NewDataSourceRepository(db)
	eventRepo := repositories.NewSecurityEventRepository(db)

	// Initialize task handler
	taskHandler := tasks.NewTaskHandler(jobRepo, dataSourceRepo, eventRepo)
	providerSyncHandler := tasks.NewProviderSyncHandler(dataSourceRepo, integrationRepo, eventRepo, taskClient, restyClient, encryptor)

	srv := asynq.NewServer(
		serverOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeIngestJob, taskHandler.HandleIngestJob)
	mux.HandleFunc(tasks.TypeSyncJob, taskHandler.HandleSyncJob)
	mux.HandleFunc(tasks.TypeQualityScanJob, taskHandler.HandleQualityScanJob)
	mux.HandleFunc(tasks.TypeValidationJob, taskHandler.HandleValidationJob)
	mux.HandleFunc(tasks.TypeProviderEventBatch, taskHandler.HandleProviderEventBatch)
	mux.HandleFunc(tasks.TypeProviderSync, providerSyncHandler.ProcessTask)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down worker...")
		srv.Shutdown()
	}()

	// Start worker
	log.Printf("Starting worker with Redis at %s", redisRef)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run worker: %v", err)
	}
}
