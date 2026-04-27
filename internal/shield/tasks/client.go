package tasks

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type TaskClient struct {
	client *asynq.Client
}

func NewTaskClient(redisAddr string) *TaskClient {
	return &TaskClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
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

func (c *TaskClient) Close() error {
	return c.client.Close()
}
