package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/providers"
	"sage-backend/internal/shield/repositories"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
)

// Task type constants are now defined in client.go
// const (
// 	TypeIngestJob          = "ingest:job"
// 	TypeSyncJob            = "sync:job"
// 	TypeQualityScanJob     = "quality:scan:job"
// 	TypeValidationJob      = "validation:job"
// 	TypeProviderEventBatch = "provider:event-batch"
// 	TypeProviderSync       = "provider:sync"
// )

type TaskHandler struct {
	jobRepo         repositories.IngestionJobRepositoryInt
	dataSourceRepo  repositories.DataSourceRepositoryInt
	eventRepo       repositories.SecurityEventRepositoryInt
	integrationRepo repositories.IntegrationRepositoryInt
	taskClient      *TaskClient
	client          *resty.Client
	encryptor       crypto.Encryptor
}

func NewTaskHandler(
	jobRepo repositories.IngestionJobRepositoryInt,
	dataSourceRepo repositories.DataSourceRepositoryInt,
	eventRepo repositories.SecurityEventRepositoryInt,
	integrationRepo repositories.IntegrationRepositoryInt,
	taskClient *TaskClient,
	client *resty.Client,
	encryptor crypto.Encryptor,
) *TaskHandler {
	return &TaskHandler{
		jobRepo:         jobRepo,
		dataSourceRepo:  dataSourceRepo,
		eventRepo:       eventRepo,
		integrationRepo: integrationRepo,
		taskClient:      taskClient,
		client:          client,
		encryptor:       encryptor,
	}
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

	// Update job status to running
	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = models.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	// Process ingestion
	if job.SourceID != nil {
		source, err := h.dataSourceRepo.GetDataSourceByID(ctx, *job.SourceID, job.OrganizationID)
		if err != nil {
			return err
		}

		// Simulate ingestion
		job.EventsProcessed = 100
		lastSync := time.Now()
		source.LastSyncAt = &lastSync
		source.TotalEvents += job.EventsProcessed
		source.EventsToday += job.EventsProcessed

		if err := h.dataSourceRepo.UpdateDataSource(ctx, source); err != nil {
			return err
		}
	}

	// Update job status to completed
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = models.JobStatusCompleted
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

	// Update job status
	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = models.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	// Simulate scan completion
	job.EventsProcessed = 1
	job.Status = models.JobStatusCompleted
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

	// Update job status
	job, err := h.jobRepo.GetJobByID(ctx, payload.JobID, payload.OrganizationID)
	if err != nil {
		return err
	}

	now := time.Now()
	job.Status = models.JobStatusRunning
	job.StartedAt = &now
	if err := h.jobRepo.UpdateJobStatus(ctx, job.ID, job.OrganizationID, job.Status, job.EventsProcessed, job.EventsFailed, job.ErrorMessage); err != nil {
		return err
	}

	// Simulate validation
	job.EventsProcessed = 1
	job.Status = models.JobStatusCompleted
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
		Events         []models.CreateRawEventResponse `json:"events"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal provider event batch: %w", err)
	}

	if len(payload.Events) == 0 {
		return nil
	}

	var security_events []*models.SecurityEvent
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
		security_events = append(security_events, &models.SecurityEvent{
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

	// get datasource
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

	// get encrypted credentials
	creds, err := h.integrationRepo.GetCredentialsByIntegration(
		ctx,
		source.ID.String(),
	)

	if err != nil {
		return err
	}

	if len(creds) == 0 {
		return fmt.Errorf("no credentials found for source %s (provider=%s); verify integration_credentials links to data_sources.id and migration 000034 has been applied", source.ID, providerName)
	}

	// decrypt credentials
	decryptedCreds := make(map[string]string)

	for _, cred := range creds {
		value, err :=
			h.encryptor.Decrypt(
				cred.EncryptedValue,
			)

		if err != nil {
			return err
		}

		decryptedCreds[cred.Key] = value
	}
	fmt.Printf("decrypted credentials: %v\n", decryptedCreds)

	checkpoint := &models.Checkpoint{
		LastCheckpoint:   source.LastCheckpoint,
		LastCheckpointAt: source.LastCheckpointAt,
	}

	// build provider
	provider, err :=
		providers.LaunchProviderSync(
			providerName,
			decryptedCreds,
			checkpoint,
			h.client,
		)

	if err != nil {
		log.Printf("Provider sync failed for source %s provider=%s: %v", source.ID, providerName, err)
		return err
	}

	// collect events
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
		log.Printf("failed to persist events for source %s: %w", source.ID, err)
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
	events []models.NormalizedEvent,
	orgID uuid.UUID,
	sourceID uuid.UUID,
) ([]models.CreateRawEventResponse, *time.Time, error) {
	var latest time.Time

	for _, ev := range events {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
	}

	chunk := 100
	var allResults []models.CreateRawEventResponse
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
func latestEventCheckpoint(events []models.RawEvent) string {
	var latest time.Time
	for _, event := range events {
		if event.EventTimeStamp.After(latest) {
			latest = event.EventTimeStamp
		}
	}

	if latest.IsZero() {
		latest = time.Now().UTC()
	}

	return latest.UTC().Format(time.RFC3339)
}
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
