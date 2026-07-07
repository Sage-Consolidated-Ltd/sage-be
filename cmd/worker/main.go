package main

import (
	"log"
	"os"
	"os/signal"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/ai_detector"
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
	aiDetectorHTTPClient := resty.New()

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

	s3Client, err := s3.NewClient(s3.S3Config{
		Region:          cfg.BaseConfig.S3Region,
		Bucket:          cfg.BaseConfig.S3Bucket,
		AccessKeyID:     cfg.BaseConfig.S3AccessKey,
		SecretAccessKey: cfg.BaseConfig.S3SecretKey,
		PresignExpiry:   24 * 60,
		MaxFileSizeMB:   20 * 1024 * 1024,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	uploader := s3.NewUploader(s3Client)
	aiDetectorBaseURL := strings.TrimSpace(cfg.DetectorAIBaseURL)
	aiDetectorToken := strings.TrimSpace(cfg.DetectorAIAuthToken)
	aiDetectorClient := ai_detector.NewAIDetectorClient(aiDetectorBaseURL, aiDetectorToken, aiDetectorHTTPClient)
	// Initialize repositories
	jobRepo := repositories.NewIngestionJobRepository(db)
	integrationRepo := repositories.NewIntegrationRepository(db)
	dataSourceRepo := repositories.NewDataSourceRepository(db)
	eventRepo := repositories.NewSecurityEventRepository(db)
	parserRepo := repositories.NewParserRepository(db)
	logUploadRepo := repositories.NewLogUploadRepository(db)
	analysisRepo := repositories.NewAnalysisRepository(db)
	threatDetector := ai_detector.NewThreatDetector(aiDetectorClient, uploader, logUploadRepo, analysisRepo)

	// Initialize task handler
	taskHandler := tasks.NewTaskHandler(jobRepo, dataSourceRepo, eventRepo, integrationRepo, parserRepo, taskClient, restyClient, encryptor, threatDetector)

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
	mux.HandleFunc(tasks.TypeProviderSync, taskHandler.HandleProviderSync)
	mux.HandleFunc(tasks.TypeSubmitLogFileForAnalysis, taskHandler.HandleSubmitLogFileForAnalysis)

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
