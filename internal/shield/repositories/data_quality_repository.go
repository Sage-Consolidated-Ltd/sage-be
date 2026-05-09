package repositories

import (
	"context"
	"database/sql"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
)

type DataQualityRepositoryInt interface {
	CreateScan(ctx context.Context, scan *models.DataQualityScan) error
	UpdateScan(ctx context.Context, scan *models.DataQualityScan) error
	GetScanByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.DataQualityScan, error)
	ListScans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]*models.DataQualityScan, error)
	CreateSourceMetric(ctx context.Context, metric *models.DataQualitySourceMetric) error
	GetSourceMetricsByScan(ctx context.Context, scanID uuid.UUID, orgID uuid.UUID) ([]*models.DataQualitySourceMetric, error)
	CreateSuggestion(ctx context.Context, suggestion *models.DataQualitySuggestion) error
	UpdateSuggestionStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.SuggestionStatus) error
	GetSuggestions(ctx context.Context, orgID uuid.UUID, sourceID, parserID *uuid.UUID, status models.SuggestionStatus) ([]*models.DataQualitySuggestion, error)
}

type DataQualityRepository struct {
	db *db.DB
}

func NewDataQualityRepository(db *db.DB) DataQualityRepositoryInt {
	return &DataQualityRepository{db: db}
}

const (
	CREATE_SCAN = `
		INSERT INTO data_quality_scans (
			organization_id, status, quality_score, parsing_errors,
			missing_fields_percentage, duplicate_events_percentage, unmapped_logs_count,
			ai_summary, started_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at
	`
	UPDATE_SCAN = `
		UPDATE data_quality_scans SET
			status = COALESCE($3, status),
			quality_score = COALESCE($4, quality_score),
			parsing_errors = COALESCE($5, parsing_errors),
			missing_fields_percentage = COALESCE($6, missing_fields_percentage),
			duplicate_events_percentage = COALESCE($7, duplicate_events_percentage),
			unmapped_logs_count = COALESCE($8, unmapped_logs_count),
			ai_summary = COALESCE($9, ai_summary),
			completed_at = COALESCE($10, completed_at),
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`
	GET_SCAN   = `SELECT * FROM data_quality_scans WHERE id = $1 AND organization_id = $2`
	LIST_SCANS = `
		SELECT * FROM data_quality_scans
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	COUNT_SCANS = `SELECT COUNT(*) FROM data_quality_scans WHERE organization_id = $1`

	CREATE_METRIC = `
		INSERT INTO data_quality_source_metrics (
			organization_id, scan_id, source_id, parsing_errors,
			missing_fields_percentage, unmapped_events, duplicate_percentage, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at
	`
	GET_METRICS_BY_SCAN = `
		SELECT * FROM data_quality_source_metrics
		WHERE scan_id = $1 AND organization_id = $2
	`

	CREATE_SUGGESTION = `
		INSERT INTO data_quality_suggestions (
			organization_id, source_id, parser_id, summary, recommendation,
			suggested_fix, confidence, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at
	`
	UPDATE_SUGGESTION_STATUS = `UPDATE data_quality_suggestions SET status = $2, applied_at = $3 WHERE id = $1 AND organization_id = $4`
	GET_SUGGESTIONS          = `
		SELECT * FROM data_quality_suggestions
		WHERE organization_id = $1
			AND ($2::uuid IS NULL OR source_id = $2)
			AND ($3::uuid IS NULL OR parser_id = $3)
			AND ($4::varchar IS NULL OR status = $4)
		ORDER BY created_at DESC
	`
)

func (r *DataQualityRepository) CreateScan(ctx context.Context, scan *models.DataQualityScan) error {
	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_SCAN,
		scan.OrganizationID, scan.Status, scan.QualityScore, scan.ParsingErrors,
		scan.MissingFieldsPercentage, scan.DuplicateEventsPercentage, scan.UnmappedLogsCount,
		scan.AISummary, scan.StartedAt, scan.CompletedAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	scan.ID = id
	scan.CreatedAt = createdAt
	return nil
}

func (r *DataQualityRepository) UpdateScan(ctx context.Context, scan *models.DataQualityScan) error {
	result, err := r.db.ExecContext(
		ctx, UPDATE_SCAN,
		scan.ID, scan.OrganizationID, scan.Status, scan.QualityScore, scan.ParsingErrors,
		scan.MissingFieldsPercentage, scan.DuplicateEventsPercentage, scan.UnmappedLogsCount,
		scan.AISummary, scan.CompletedAt,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("SCAN NOT FOUND")
	}
	return nil
}

func (r *DataQualityRepository) GetScanByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.DataQualityScan, error) {
	var scan models.DataQualityScan
	err := r.db.GetContext(ctx, &scan, GET_SCAN, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("SCAN NOT FOUND")
		}
		return nil, err
	}
	return &scan, nil
}

func (r *DataQualityRepository) ListScans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]*models.DataQualityScan, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize
	var scans []*models.DataQualityScan
	err := r.db.SelectContext(ctx, &scans, LIST_SCANS, orgID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return scans, nil
}

func (r *DataQualityRepository) CreateSourceMetric(ctx context.Context, metric *models.DataQualitySourceMetric) error {
	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_METRIC,
		metric.OrganizationID, metric.ScanID, metric.SourceID, metric.ParsingErrors,
		metric.MissingFieldsPercentage, metric.UnmappedEvents, metric.DuplicatePercentage, metric.Status,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	metric.ID = id
	metric.CreatedAt = createdAt
	return nil
}

func (r *DataQualityRepository) GetSourceMetricsByScan(ctx context.Context, scanID uuid.UUID, orgID uuid.UUID) ([]*models.DataQualitySourceMetric, error) {
	var metrics []*models.DataQualitySourceMetric
	err := r.db.SelectContext(ctx, &metrics, GET_METRICS_BY_SCAN, scanID, orgID)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *DataQualityRepository) CreateSuggestion(ctx context.Context, suggestion *models.DataQualitySuggestion) error {
	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_SUGGESTION,
		suggestion.OrganizationID, suggestion.SourceID, suggestion.ParserID,
		suggestion.Summary, suggestion.Recommendation, suggestion.SuggestedFix,
		suggestion.Confidence, suggestion.Status,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	suggestion.ID = id
	suggestion.CreatedAt = createdAt
	return nil
}

func (r *DataQualityRepository) UpdateSuggestionStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.SuggestionStatus) error {
	var appliedAt *time.Time
	if status == models.SuggestionStatusApplied {
		now := time.Now()
		appliedAt = &now
	}
	_, err := r.db.ExecContext(ctx, UPDATE_SUGGESTION_STATUS, id, status, appliedAt, orgID)
	return err
}

func (r *DataQualityRepository) GetSuggestions(ctx context.Context, orgID uuid.UUID, sourceID, parserID *uuid.UUID, status models.SuggestionStatus) ([]*models.DataQualitySuggestion, error) {
	var suggestions []*models.DataQualitySuggestion
	source := ""
	if sourceID != nil {
		source = sourceID.String()
	}
	parser := ""
	if parserID != nil {
		parser = parserID.String()
	}
	statusStr := ""
	if status != "" {
		statusStr = string(status)
	}
	// Build query with optional filters using OR conditions
	query := GET_SUGGESTIONS + " ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &suggestions, query, orgID, source, parser, statusStr)
	if err != nil {
		return nil, err
	}
	return suggestions, nil
}
