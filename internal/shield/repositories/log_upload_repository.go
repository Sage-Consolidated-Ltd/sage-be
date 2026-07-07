package repositories

import (
	"context"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/models"
)



type LogUploadRepositoryInt interface {
	CreatePending(ctx context.Context, params models.CreateLogFileParams) (*models.LogFile, error)
	Confirm(ctx context.Context, params models.ConfirmLogFileParams) (*models.LogFile, error)
	GetByS3Key(ctx context.Context, s3Key string) (*models.LogFile, error)
	MarkSubmitted(ctx context.Context, s3Key string) error
	MarkFailed(ctx context.Context, s3Key string, reason string) error
}

type LogUploadRepository struct {
	db *db.DB
}

func NewLogUploadRepository(db *db.DB) LogUploadRepositoryInt {
	return &LogUploadRepository{db: db}
}

func (r *LogUploadRepository) CreatePending(ctx context.Context, params models.CreateLogFileParams) (*models.LogFile, error) {
	const q = `
		INSERT INTO log_files (user_id, organization_id, s3_key, file_class, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING *`

	var lf models.LogFile
	err := r.db.QueryRowxContext(ctx, q,
		params.UserID,
		params.OrganizationID,
		params.S3Key,
		params.FileClass,
	).StructScan(&lf)
	if err != nil {
		return nil, err
	}

	return &lf, nil
}

func (r *LogUploadRepository) Confirm(ctx context.Context, params models.ConfirmLogFileParams) (*models.LogFile, error) {
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

	var lf models.LogFile
	err := r.db.QueryRowxContext(ctx, q,
		params.S3Key,
		params.SourceType,
		params.SourceID,
		params.Description,
		params.Category,
		params.AppOrContext,
	).StructScan(&lf)
	if err != nil {
		return nil, err
	}

	return &lf, nil
}

func (r *LogUploadRepository) GetByS3Key(ctx context.Context, s3Key string) (*models.LogFile, error) {
	const q = `SELECT * FROM log_files WHERE s3_key = $1`

	var lf models.LogFile
	err := r.db.QueryRowxContext(ctx, q, s3Key).StructScan(&lf)
	if err != nil {
		return nil, err
	}

	return &lf, nil
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