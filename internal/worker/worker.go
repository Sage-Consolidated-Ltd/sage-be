package worker

import (
	"context"
	"fmt"
	"os"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shield/adapters/outbound/postgres"
	"sage-backend/internal/shield/ai_detector"
	"sage-backend/internal/shield/tasks"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	config *config.WorkerConfig
	server *asynq.Server
	mux    *asynq.ServeMux
}

func New() (*Worker, error) {
	cfg := config.SetupWorker()

	database, err := db.ConnectDB(&cfg.BaseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Init(&cfg.BaseConfig)

	redisOpt, err := parseRedisOpt(cfg.RedisDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	encryptor, err := crypto.NewAESEncryptor(cfg.AppEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
	}

	restyClient := resty.New()
	taskClient := tasks.NewTaskClient(cfg.RedisDbUrl)

	var s3Uploader *s3.Uploader
	s3Client, err := s3.NewClient(context.Background(), cfg.S3Bucket, cfg.S3Region)
	if err == nil && s3Client != nil {
		s3Uploader = s3.NewUploader(s3Client)
	}

	eventRepo := postgres.NewSecurityEventRepository(database)
	dataSourceRepo := postgres.NewDataSourceRepository(database)
	jobRepo := postgres.NewIngestionJobRepository(database)
	integrationRepo := postgres.NewIntegrationRepository(database)
	logUploadRepo := postgres.NewLogUploadRepository(database)
	parsedLogRepo := postgres.NewParsedLogRepository(database)
	analysisRepo := postgres.NewAnalysisRepository(database)

	aiDetectorBaseURL := os.Getenv("DETECTOR_AI_BASE_URL")
	aiDetectorToken := os.Getenv("DETECTOR_AI_AUTH_TOKEN")
	aiDetectorHTTPClient := resty.New()
	aiDetectorClient := ai_detector.NewAIDetectorClient(aiDetectorBaseURL, aiDetectorToken, aiDetectorHTTPClient)
	threatDetector := ai_detector.NewThreatDetector(aiDetectorClient, s3Uploader, logUploadRepo, analysisRepo)

	taskHandler := tasks.NewTaskHandler(
		jobRepo,
		dataSourceRepo,
		eventRepo,
		integrationRepo,
		taskClient,
		restyClient,
		encryptor,
		threatDetector,
		s3Uploader,
		parsedLogRepo,
	)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeIngestJob, taskHandler.HandleIngestJob)
	mux.HandleFunc(tasks.TypeSyncJob, taskHandler.HandleSyncJob)
	mux.HandleFunc(tasks.TypeQualityScanJob, taskHandler.HandleQualityScanJob)
	mux.HandleFunc(tasks.TypeValidationJob, taskHandler.HandleValidationJob)
	mux.HandleFunc(tasks.TypeProviderEventBatch, taskHandler.HandleProviderEventBatch)
	mux.HandleFunc(tasks.TypeProviderSync, taskHandler.HandleProviderSync)
	mux.HandleFunc(tasks.TypeSubmitLogFileForProcessing, taskHandler.HandleProcessLogFile)

	return &Worker{
		config: cfg,
		server: server,
		mux:    mux,
	}, nil
}

func (w *Worker) Run() error {
	logger.Info("Starting background worker process...")
	if err := w.server.Run(w.mux); err != nil {
		return fmt.Errorf("worker server error: %w", err)
	}
	return nil
}

func parseRedisOpt(redisURL string) (asynq.RedisClientOpt, error) {
	if parsed, err := redis.ParseURL(redisURL); err == nil {
		return asynq.RedisClientOpt{
			Addr:     parsed.Addr,
			Password: parsed.Password,
			DB:       parsed.DB,
		}, nil
	}
	return asynq.RedisClientOpt{
		Addr: redisURL,
	}, nil
}
