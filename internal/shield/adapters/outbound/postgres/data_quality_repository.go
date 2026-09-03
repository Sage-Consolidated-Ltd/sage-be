package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/adapters/outbound/postgres/models"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type DataQualityRepository struct {
	db *db.DB
}

func NewDataQualityRepository(db *db.DB) outbound.DataQualityRepository {
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

func (r *DataQualityRepository) CreateScan(ctx context.Context, scan *domain.DataQualityScan) error {
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

func (r *DataQualityRepository) UpdateScan(ctx context.Context, scan *domain.DataQualityScan) error {
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

func (r *DataQualityRepository) GetScanByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataQualityScan, error) {
	var dto models.DataQualityScanDTO
	err := r.db.GetContext(ctx, &dto, GET_SCAN, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("SCAN NOT FOUND")
		}
		return nil, err
	}
	return dto.ToDomain(), nil
}

func (r *DataQualityRepository) ListScans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]*domain.DataQualityScan, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize
	var dtos []*models.DataQualityScanDTO
	err := r.db.SelectContext(ctx, &dtos, LIST_SCANS, orgID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	scans := make([]*domain.DataQualityScan, 0, len(dtos))
	for _, dto := range dtos {
		scans = append(scans, dto.ToDomain())
	}
	return scans, nil
}

func (r *DataQualityRepository) CreateSourceMetric(ctx context.Context, metric *domain.DataQualitySourceMetric) error {
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

func (r *DataQualityRepository) GetSourceMetricsByScan(ctx context.Context, scanID uuid.UUID, orgID uuid.UUID) ([]*domain.DataQualitySourceMetric, error) {
	var dtos []*models.DataQualitySourceMetricDTO
	err := r.db.SelectContext(ctx, &dtos, GET_METRICS_BY_SCAN, scanID, orgID)
	if err != nil {
		return nil, err
	}

	metrics := make([]*domain.DataQualitySourceMetric, 0, len(dtos))
	for _, dto := range dtos {
		metrics = append(metrics, dto.ToDomain())
	}
	return metrics, nil
}

func (r *DataQualityRepository) CreateSuggestion(ctx context.Context, suggestion *domain.DataQualitySuggestion) error {
	var id uuid.UUID
	var createdAt time.Time
	fixJSON, err := json.Marshal(suggestion.SuggestedFix)
	if err != nil {
		return err
	}
	err = r.db.QueryRowContext(
		ctx, CREATE_SUGGESTION,
		suggestion.OrganizationID, suggestion.SourceID, suggestion.ParserID,
		suggestion.Summary, suggestion.Recommendation, fixJSON, suggestion.Confidence, suggestion.Status,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	suggestion.ID = id
	suggestion.CreatedAt = createdAt
	return nil
}

func (r *DataQualityRepository) UpdateSuggestionStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.SuggestionStatus) error {
	var appliedAt *time.Time
	if status == domain.SuggestionStatusApplied {
		now := time.Now()
		appliedAt = &now
	}
	_, err := r.db.ExecContext(ctx, UPDATE_SUGGESTION_STATUS, id, status, appliedAt, orgID)
	return err
}

func (r *DataQualityRepository) GetSuggestions(ctx context.Context, orgID uuid.UUID, sourceID, parserID *uuid.UUID, status domain.SuggestionStatus) ([]*domain.DataQualitySuggestion, error) {
	var dtos []*models.DataQualitySuggestionDTO
	var statusVal *string
	if status != "" {
		s := string(status)
		statusVal = &s
	}
	query := GET_SUGGESTIONS
	err := r.db.SelectContext(ctx, &dtos, query, orgID, sourceID, parserID, statusVal)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*domain.DataQualitySuggestion, 0, len(dtos))
	for _, dto := range dtos {
		suggestions = append(suggestions, dto.ToDomain())
	}
	return suggestions, nil
}
