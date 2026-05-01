package repositories

import (
	"context"
	"database/sql"
	"time"

	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IngestionJobRepositoryInt interface {
	CreateJob(ctx context.Context, job *models.IngestionJob) error
	GetJobByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.IngestionJob, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.JobStatus, eventsProcessed, eventsFailed int64, errMsg *string) error
	ListJobs(ctx context.Context, orgID uuid.UUID, jobType models.JobType, status models.JobStatus, page, pageSize int) ([]*models.IngestionJob, int, error)
}

type IngestionJobRepository struct {
	db *sqlx.DB
}

func NewIngestionJobRepository(db *sqlx.DB) IngestionJobRepositoryInt {
	return &IngestionJobRepository{db: db}
}

const (
	CREATE_JOB = `
		INSERT INTO ingestion_jobs (
			organization_id, source_id, status, job_type,
			metadata, started_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at, updated_at
	`
	GET_JOB           = `SELECT * FROM ingestion_jobs WHERE id = $1 AND organization_id = $2`
	UPDATE_JOB_STATUS = `
		UPDATE ingestion_jobs SET
			status = $2,
			events_processed = COALESCE($3, events_processed),
			events_failed = COALESCE($4, events_failed),
			error_message = $5,
			started_at = COALESCE($6, started_at),
			completed_at = $7,
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $8
	`
	LIST_JOBS = `
		SELECT * FROM ingestion_jobs
		WHERE organization_id = $1
			AND ($2::varchar IS NULL OR job_type = $2)
			AND ($3::varchar IS NULL OR status = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	COUNT_JOBS = `
		SELECT COUNT(*) FROM ingestion_jobs
		WHERE organization_id = $1
			AND ($2::varchar IS NULL OR job_type = $2)
			AND ($3::varchar IS NULL OR status = $3)
	`
)

func (r *IngestionJobRepository) CreateJob(ctx context.Context, job *models.IngestionJob) error {
	var id uuid.UUID
	var createdAt, updatedAt time.Time
	meta := job.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	}
	err := r.db.QueryRowContext(
		ctx, CREATE_JOB,
		job.OrganizationID, job.SourceID, job.Status, job.JobType,
		meta, job.StartedAt, job.CompletedAt,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}
	job.ID = id
	job.CreatedAt = createdAt
	job.UpdatedAt = updatedAt
	job.Metadata = meta
	return nil
}

func (r *IngestionJobRepository) GetJobByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.IngestionJob, error) {
	var job models.IngestionJob
	err := r.db.GetContext(ctx, &job, GET_JOB, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("JOB NOT FOUND")
		}
		return nil, err
	}
	return &job, nil
}

func (r *IngestionJobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.JobStatus, eventsProcessed, eventsFailed int64, errMsg *string) error {
	result, err := r.db.ExecContext(
		ctx, UPDATE_JOB_STATUS,
		id, status, eventsProcessed, eventsFailed, errMsg, time.Now(), time.Now(), orgID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("JOB NOT FOUND")
	}
	return nil
}

func (r *IngestionJobRepository) ListJobs(ctx context.Context, orgID uuid.UUID, jobType models.JobType, status models.JobStatus, page, pageSize int) ([]*models.IngestionJob, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	jobTypeStr := ""
	if jobType != "" {
		jobTypeStr = string(jobType)
	}
	statusStr := ""
	if status != "" {
		statusStr = string(status)
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_JOBS, orgID, jobTypeStr, statusStr)
	if err != nil {
		return nil, 0, err
	}

	var jobs []*models.IngestionJob
	err = r.db.SelectContext(ctx, &jobs, LIST_JOBS, orgID, jobTypeStr, statusStr, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}
