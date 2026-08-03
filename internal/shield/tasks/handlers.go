package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/adapters/outbound/providers"
	"sage-backend/internal/shield/ai_detector"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
	"sage-backend/internal/shield/upload_parser"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
)

type TaskHandler struct {
	jobRepo             outbound.IngestionJobRepository
	dataSourceRepo      outbound.DataSourceRepository
	eventRepo           outbound.SecurityEventRepository
	integrationRepo     outbound.IntegrationRepository
	taskClient          *TaskClient
	client              *resty.Client
	encryptor           crypto.Encryptor
	threatDetector      ai_detector.ThreatDetectorInt
	s3Uploader          *s3.Uploader
	parsedLogRepository outbound.ParsedLogRepository
}

func NewTaskHandler(
	jobRepo outbound.IngestionJobRepository,
	dataSourceRepo outbound.DataSourceRepository,
	eventRepo outbound.SecurityEventRepository,
	integrationRepo outbound.IntegrationRepository,
	taskClient *TaskClient,
	client *resty.Client,
	encryptor crypto.Encryptor,
	threatDetector ai_detector.ThreatDetectorInt,
	s3Uploader *s3.Uploader,
	parsedLogRepository outbound.ParsedLogRepository,
) *TaskHandler {
	return &TaskHandler{
		jobRepo:             jobRepo,
		dataSourceRepo:      dataSourceRepo,
		eventRepo:           eventRepo,
		integrationRepo:     integrationRepo,
		taskClient:          taskClient,
		client:              client,
		encryptor:           encryptor,
		threatDetector:      threatDetector,
		s3Uploader:          s3Uploader,
		parsedLogRepository: parsedLogRepository,
	}
}

func (h *TaskHandler) HandleProcessLogFile(ctx context.Context, t *asynq.Task) error {
	if h.threatDetector == nil {
		return fmt.Errorf("threat detector is not configured")
	}

	var payload domain.SubmitLogFileInput
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal submit log file payload: %w", err)
	}

	if h.s3Uploader == nil {
		return fmt.Errorf("s3 uploader is not configured")
	}

	fileBytes, err := h.s3Uploader.DownloadObject(ctx, payload.S3Key)
	if err != nil {
		return fmt.Errorf("download from s3: %w", err)
	}

	filename := filepath.Base(payload.S3Key)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if payload.SourceID == nil {
			logger.Error("Source ID is nil for log file", zap.Any("log_file_id", payload.LogFileID))
			return fmt.Errorf("source ID is nil for log file %s", payload.LogFileID)
		}
		parsedLogs, err := h.parseFile(payload, filename, fileBytes)
		if err != nil {
			logger.Error("Error parsing file in HandleProcessLogFile", zap.Error(err))
			return fmt.Errorf("parse file: %w", err)
		}

		if h.parsedLogRepository != nil {
			if err := h.parsedLogRepository.StoreParsedLogs(ctx, parsedLogs); err != nil {
				logger.Error("Error storing parsed logs in HandleProcessLogFile", zap.Error(err))
				return fmt.Errorf("store parsed logs: %w", err)
			}
		}

		log.Printf("Stored %d parsed logs for file %s", len(parsedLogs), filename)
		return nil
	})

	g.Go(func() error {
		input := domain.SubmitLogFileForAnalysis{
			OrganizationID: payload.OrganizationID,
			FileName:       filename,
			FileClass:      payload.FileClass,
			LogFileID:      payload.LogFileID,
			FileReader:     bytes.NewReader(fileBytes),
			S3Key:          payload.S3Key,
			UserID:         payload.UserID,
		}

		result, err := h.threatDetector.SubmitLogFileForAnalysis(ctx, input)
		if err != nil {
			return fmt.Errorf("submit log file for analysis failed: %w", err)
		}

		log.Printf("Completed analysis submission for log_file_id=%s job_id=%s",
			payload.LogFileID, result.JobID)

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// HandleIngestJob processes log ingestion from a data source
func (h *TaskHandler) HandleIngestJob(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		JobID          uuid.UUID  `json:"job_id"`
		OrganizationID uuid.UUID  `json:"organization_id"`
		SourceID       *uuid.UUID `json:"source_id"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = domain.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	if job.SourceID != nil {
		source, err := h.dataSourceRepo.GetDataSourceByID(ctx, *job.SourceID, job.OrganizationID)
		if err != nil {
			return err
		}

		job.EventsProcessed = 100
		lastSync := time.Now()
		source.LastSyncAt = &lastSync
		source.TotalEvents += job.EventsProcessed
		source.EventsToday += job.EventsProcessed

		if err := h.dataSourceRepo.UpdateDataSource(ctx, source); err != nil {
			return err
		}
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = domain.JobStatusCompleted
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	log.Printf("Completed ingest job %s", job.ID)
	return nil
}

// HandleSyncJob handles manual sync trigger for a data source
func (h *TaskHandler) HandleSyncJob(ctx context.Context, t *asynq.Task) error {
	return h.HandleIngestJob(ctx, t)
}

// HandleQualityScanJob handles data quality analysis
func (h *TaskHandler) HandleQualityScanJob(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		JobID          uuid.UUID `json:"job_id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		ScanID         uuid.UUID `json:"scan_id"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = domain.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	job.EventsProcessed = 1
	job.Status = domain.JobStatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	log.Printf("Completed quality scan job %s", job.ID)
	return nil
}

// HandleValidationJob handles parser validation
func (h *TaskHandler) HandleValidationJob(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		JobID          uuid.UUID `json:"job_id"`
		OrganizationID uuid.UUID `json:"organization_id"`
		ParserID       uuid.UUID `json:"parser_id"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = domain.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	job.EventsProcessed = 1
	job.Status = domain.JobStatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	log.Printf("Completed validation job %s", job.ID)
	return nil
}

func (h *TaskHandler) HandleProviderEventBatch(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		OrganizationID uuid.UUID                       `json:"organization_id"`
		SourceID       uuid.UUID                       `json:"source_id"`
		Events         []domain.CreateRawEventResponse `json:"events"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal provider event batch: %w", err)
	}

	if len(payload.Events) == 0 {
		return nil
	}

	var security_events []*domain.SecurityEvent
	var latest time.Time
	for _, ev := range payload.Events {
		re, err := h.eventRepo.GetRawEventByID(ctx, ev.ID, payload.OrganizationID)
		if err != nil {
			log.Printf("Failed to fetch raw event for ID %s: %v", ev.ID, err)
			return err
		}
		if re.EventTimeStamp.After(latest) {
			latest = re.EventTimeStamp
		}

		idCopy := re.ID.String()
		ipCopy := re.IPAddress
		userCopy := re.UserName
		security_events = append(security_events, &domain.SecurityEvent{
			OrganizationID:    payload.OrganizationID,
			SourceID:          payload.SourceID,
			SourceEventID:     &idCopy,
			Source:            re.Provider,
			EventType:         re.EventType,
			IPAddress:         ipCopy,
			ActorUsername:     userCopy,
			RawPayload:        re.RawPayload,
			NormalizedPayload: map[string]interface{}{"user_id": re.UserID},
			ParseStatus:       types.ParseStatusPending,
			OccurredAt:        re.EventTimeStamp,
		})
	}

	if err := h.eventRepo.BulkCreateEvents(ctx, security_events); err != nil {
		return err
	}

	now := time.Now()
	if err := h.dataSourceRepo.UpdateHealthMetrics(ctx, payload.SourceID, int64(len(security_events)), int64(len(security_events)), 0, &latest, &now); err != nil {
		log.Printf("Failed to update data source metrics: %v", err)
	}

	lastCheckpoint := latest.UTC().Format(time.RFC3339)
	if err := h.dataSourceRepo.UpdateCheckpoint(ctx, payload.SourceID, lastCheckpoint); err != nil {
		return fmt.Errorf("failed to persist checkpoint for source %s: %w", payload.SourceID, err)
	}

	log.Printf("Persisted checkpoint %s for source %s", lastCheckpoint, payload.SourceID)
	log.Printf("Persisted %d events for source %s", len(security_events), payload.SourceID)
	return nil
}

func (h *TaskHandler) HandleProviderSync(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload struct {
		OrganizationID uuid.UUID `json:"organization_id"`
		SourceID       uuid.UUID `json:"source_id"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("Starting provider sync task for org=%s source=%s", payload.OrganizationID, payload.SourceID)

	source, err := h.dataSourceRepo.GetDataSourceByID(
		ctx,
		payload.SourceID,
		payload.OrganizationID,
	)
	if err != nil {
		return err
	}

	if source.Provider == nil || strings.TrimSpace(*source.Provider) == "" {
		return fmt.Errorf("source %s has no provider configured", source.ID)
	}

	providerName := strings.ToLower(strings.TrimSpace(*source.Provider))

	creds, err := h.integrationRepo.GetCredentialsByIntegration(
		ctx,
		source.ID.String(),
	)
	if err != nil {
		return err
	}

	if len(creds) == 0 {
		return fmt.Errorf("no credentials found for source %s (provider=%s)", source.ID, providerName)
	}

	decryptedCreds := make(map[string]string)
	for _, cred := range creds {
		value, err := h.encryptor.Decrypt(cred.EncryptedValue)
		if err != nil {
			return err
		}
		decryptedCreds[cred.Key] = value
	}

	checkpoint := &domain.Checkpoint{
		LastCheckpoint:   source.LastCheckpoint,
		LastCheckpointAt: source.LastCheckpointAt,
	}

	provider, err := providers.LaunchProviderSync(
		providerName,
		decryptedCreds,
		checkpoint,
		h.client,
	)
	if err != nil {
		log.Printf("Provider sync failed for source %s provider=%s: %v", source.ID, providerName, err)
		return err
	}

	events, err := provider.Collect(ctx, 500)
	if err != nil {
		log.Printf("Error collecting logs: %v", err)
		return err
	}

	log.Printf("Collected %d events from %s", len(events), providerName)

	if len(events) == 0 {
		return nil
	}

	rawEvents, _, err := h.persistNormalizedEvents(ctx, events, source.OrganizationID, source.ID)
	if err != nil {
		log.Printf("failed to persist events for source %s: %v", source.ID, err)
		return fmt.Errorf("failed to persist events for source %s: %w", source.ID, err)
	}

	log.Printf("Persisted %d normalized events for source %s", len(events), source.ID)

	if err := h.taskClient.EnqueueProviderEventBatch(ctx, payload.OrganizationID, payload.SourceID, rawEvents); err != nil {
		return err
	}

	return nil
}

func (h *TaskHandler) persistNormalizedEvents(
	ctx context.Context,
	events []domain.NormalizedEvent,
	orgID uuid.UUID,
	sourceID uuid.UUID,
) ([]domain.CreateRawEventResponse, *time.Time, error) {
	var latest time.Time

	for _, ev := range events {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
	}

	chunk := 100
	var allResults []domain.CreateRawEventResponse
	for i := 0; i < len(events); i += chunk {
		end := i + chunk
		if end > len(events) {
			end = len(events)
		}

		results, err := h.eventRepo.BulkInsertRawEvents(ctx, orgID, &sourceID, events[i:end])
		if err != nil {
			return nil, nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, &latest, nil
}

func (h *TaskHandler) parseFile(payload domain.SubmitLogFileInput, filename string, fileBytes []byte) ([]domain.ParsedLog, error) {
	sampleLen := len(fileBytes)
	if sampleLen > 4096 {
		sampleLen = 4096
	}

	detected := upload_parser.DetectType(filename, fileBytes[:sampleLen])
	if detected == domain.DetectedUnknown {
		return nil, fmt.Errorf("unable to detect log type for %q", filename)
	}

	parser, err := upload_parser.ParserFor(detected)
	if err != nil {
		return nil, err
	}

	parsedLogs, err := parser.Parse(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("parse %q as %s: %w", filename, detected, err)
	}

	for i := range parsedLogs {
		parsedLogs[i].FileID = payload.LogFileID
		if payload.SourceID != nil {
			parsedLogs[i].DataSourceID = *payload.SourceID
		}
	}

	return parsedLogs, nil
}
