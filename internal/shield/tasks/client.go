package tasks

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const (
	TypeProviderSync       = "provider:sync"
	TypeProviderEventBatch = "provider:event-batch"
	TypeIngestJob          = "ingest:job"
	TypeSyncJob            = "sync:job"
	TypeQualityScanJob     = "quality:scan:job"
	TypeValidationJob      = "validation:job"
	TypeSubmitLogFileForAnalysis = "analysis:submit-log-file"
)

type TaskClient struct {
	client *asynq.Client
}

func NewTaskClient(redisRef string) *TaskClient {
	opt := asynq.RedisClientOpt{Addr: redisRef}
	if strings.Contains(redisRef, "://") {
		if parsed, err := redis.ParseURL(redisRef); err == nil {
			opt = asynq.RedisClientOpt{
				Addr:     parsed.Addr,
				Password: parsed.Password,
				DB:       parsed.DB,
			}
		}
	}

	return &TaskClient{
		client: asynq.NewClient(opt),
	}
}

func (c *TaskClient) EnqueueProviderSync(ctx context.Context, orgID uuid.UUID, sourceID uuid.UUID) error {
	payload, err := json.Marshal(
		map[string]interface{}{
			"organization_id": orgID,
			"source_id":       sourceID,
		},
	)

	if err != nil {
		return err
	}

	task := asynq.NewTask(
		TypeProviderSync,
		payload,
	)

	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(5),
		asynq.Queue("default"),
	)

	return err
}

func (c *TaskClient) EnqueueProviderEventBatch(
	ctx context.Context,
	orgID uuid.UUID,
	sourceID uuid.UUID,
	events []models.CreateRawEventResponse,
) error {
	batchSize := 100
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		chunk := events[i:end]

		payload, err := json.Marshal(map[string]interface{}{
			"organization_id": orgID,
			"source_id":       sourceID,
			"events":          chunk,
		})
		if err != nil {
			return err
		}

		task := asynq.NewTask(TypeProviderEventBatch, payload)

		if _, err := c.client.Enqueue(task, asynq.Queue("default")); err != nil {
			return err
		}
	}

	return nil
}

// EnqueueIngestJob enqueues an ingestion job
func (c *TaskClient) EnqueueIngestJob(ctx context.Context, jobID, orgID uuid.UUID, sourceID *uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"job_id":          jobID,
		"organization_id": orgID,
		"source_id":       sourceID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeIngestJob, payload)
	_, err = c.client.Enqueue(task)
	return err
}

// EnqueueSyncJob enqueues a sync job
func (c *TaskClient) EnqueueSyncJob(ctx context.Context, jobID, orgID uuid.UUID, sourceID *uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"job_id":          jobID,
		"organization_id": orgID,
		"source_id":       sourceID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeSyncJob, payload)
	_, err = c.client.Enqueue(task)
	return err
}

// EnqueueQualityScanJob enqueues a quality scan job
func (c *TaskClient) EnqueueQualityScanJob(ctx context.Context, jobID, orgID, scanID uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"job_id":          jobID,
		"organization_id": orgID,
		"scan_id":         scanID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeQualityScanJob, payload)
	_, err = c.client.Enqueue(task)
	return err
}

// EnqueueValidationJob enqueues a validation job
func (c *TaskClient) EnqueueValidationJob(ctx context.Context, jobID, orgID, parserID uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"job_id":          jobID,
		"organization_id": orgID,
		"parser_id":       parserID,
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeValidationJob, payload)
	_, err = c.client.Enqueue(task)
	return err
}

// EnqueueSubmitLogFileForAnalysis enqueues uploaded log analysis for worker processing.
func (c *TaskClient) EnqueueSubmitLogFileForAnalysis(ctx context.Context, input models.SubmitLogFileInput) error {
	log.Printf("Enqueuing log file for analysis: LogFileID=%s, OrganizationID=%s, S3Key=%s", input.LogFileID, input.OrganizationID, input.S3Key)
	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeSubmitLogFileForAnalysis, payload)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(5), asynq.Queue("default"))
	return err
}

func (c *TaskClient) Close() error {
	return c.client.Close()
}
