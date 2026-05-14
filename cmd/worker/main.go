package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/repositories"
	"sage-backend/internal/shield/tasks"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	taskClient := tasks.NewTaskClient(redisAddr)
	defer taskClient.Close()

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
	providerSyncHandler := tasks.NewProviderSyncHandler(dataSourceRepo, integrationRepo, taskClient, restyClient, encryptor)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
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
	log.Printf("Starting worker with Redis at %s", redisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run worker: %v", err)
	}
}
