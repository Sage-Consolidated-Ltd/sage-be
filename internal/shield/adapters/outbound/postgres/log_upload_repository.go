package postgres

import (
	"context"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/adapters/outbound/postgres/models"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
)

type LogUploadRepository struct {
	db *db.DB
}

func NewLogUploadRepository(database *db.DB) outbound.LogUploadRepository {
	return &LogUploadRepository{db: database}
}

func (r *LogUploadRepository) CreatePending(ctx context.Context, params domain.CreateLogFileParams) (*domain.LogFile, error) {
	const q = `
		INSERT INTO log_files (user_id, organization_id, s3_key, file_class, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING *`

	var dto models.LogFileDTO
	err := r.db.QueryRowxContext(ctx, q,
		params.UserID,
		params.OrganizationID,
		params.S3Key,
		params.FileClass,
	).StructScan(&dto)
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(), nil
}

func (r *LogUploadRepository) Confirm(ctx context.Context, params domain.ConfirmLogFileParams) (*domain.LogFile, error) {
	const q = `
		UPDATE log_files
		SET
			status      = 'uploaded',
			source_type = $2,
			source_id   = $3,
			description = $4,
			category = $5,
			app_or_context = $6,
			updated_at  = now()
		WHERE s3_key = $1
		  AND status = 'pending'
		RETURNING *`

	var dto models.LogFileDTO
	err := r.db.QueryRowxContext(ctx, q,
		params.S3Key,
		params.SourceType,
		params.SourceID,
		params.Description,
		params.Category,
		params.AppOrContext,
	).StructScan(&dto)
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(), nil
}

func (r *LogUploadRepository) GetByS3Key(ctx context.Context, s3Key string) (*domain.LogFile, error) {
	const q = `SELECT * FROM log_files WHERE s3_key = $1`

	var dto models.LogFileDTO
	err := r.db.QueryRowxContext(ctx, q, s3Key).StructScan(&dto)
	if err != nil {
		return nil, err
	}

	return dto.ToDomain(), nil
}

func (r *LogUploadRepository) MarkSubmitted(ctx context.Context, s3Key string) error {
	const q = `
		UPDATE log_files
		SET status = 'submitted', updated_at = now()
		WHERE s3_key = $1`

	_, err := r.db.ExecContext(ctx, q, s3Key)
	return err
}

func (r *LogUploadRepository) MarkFailed(ctx context.Context, s3Key string, reason string) error {
	const q = `
		UPDATE log_files
		SET status = 'failed', error_message = $2, updated_at = now()
		WHERE s3_key = $1`

	_, err := r.db.ExecContext(ctx, q, s3Key, reason)
	return err
}
