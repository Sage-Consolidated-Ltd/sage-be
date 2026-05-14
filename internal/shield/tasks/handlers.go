package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
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
	jobRepo        repositories.IngestionJobRepositoryInt
	dataSourceRepo repositories.DataSourceRepositoryInt
	eventRepo      repositories.SecurityEventRepositoryInt
}

func NewTaskHandler(
	jobRepo repositories.IngestionJobRepositoryInt,
	dataSourceRepo repositories.DataSourceRepositoryInt,
	eventRepo repositories.SecurityEventRepositoryInt,
) *TaskHandler {
	return &TaskHandler{
		jobRepo:        jobRepo,
		dataSourceRepo: dataSourceRepo,
		eventRepo:      eventRepo,
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
		OrganizationID uuid.UUID                `json:"organization_id"`
		SourceID       uuid.UUID                `json:"source_id"`
		Provider       string                   `json:"provider"`
		Events         []models.NormalizedEvent `json:"events"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal provider event batch: %w", err)
	}

	if len(payload.Events) == 0 {
		return nil
	}

	var events []*models.SecurityEvent
	var latest time.Time
	for _, ev := range payload.Events {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
		idCopy := ev.ID
		ipCopy := ev.IPAddress
		userCopy := ev.UserName
		events = append(events, &models.SecurityEvent{
			OrganizationID:    payload.OrganizationID,
			SourceID:          payload.SourceID,
			SourceEventID:     &idCopy,
			Source:            payload.Provider,
			EventType:         ev.EventType,
			IPAddress:         &ipCopy,
			ActorUsername:     &userCopy,
			RawPayload:        ev.Raw,
			NormalizedPayload: map[string]interface{}{"user_id": ev.UserID},
			ParseStatus:       types.ParseStatusPending,
			OccurredAt:        ev.Timestamp,
		})
	}

	if err := h.eventRepo.BulkCreateEvents(ctx, events); err != nil {
		return err
	}

	now := time.Now()
	if err := h.dataSourceRepo.UpdateHealthMetrics(ctx, payload.SourceID, int64(len(events)), int64(len(events)), 0, &latest, &now); err != nil {
		log.Printf("Failed to update data source metrics: %v", err)
	}

	log.Printf("Persisted %d events for source %s", len(events), payload.SourceID)
	return nil
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
