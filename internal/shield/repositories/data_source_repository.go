package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
)

type DataSourceRepositoryInt interface {
	CreateDataSource(ctx context.Context, ds *models.DataSource) error
	UpdateDataSource(ctx context.Context, ds *models.DataSource) error
	GetDataSourceByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.DataSource, error)
	ListDataSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.DataSource, int, error)
	UpdateHealthMetrics(ctx context.Context, id uuid.UUID, eventsToday, totalEvents, errorCount int64, lastEventAt, lastSyncAt *time.Time) error
	IncrementEventsToday(ctx context.Context, id uuid.UUID) error
	ResetDailyCounts(ctx context.Context) error
	DisconnectDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DeleteDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	GetAggregatedHealth(ctx context.Context, orgID uuid.UUID) (totalEvents, activeSources, delayedSources, errorSources int64, err error)
	GetSourcesWithIssues(ctx context.Context, orgID uuid.UUID) ([]*models.DataSource, error)
	ListAllActiveDataSources(ctx context.Context) ([]*models.DataSource, error)
	GetCheckpoint(ctx context.Context, id uuid.UUID) (*string, error)
	UpdateCheckpoint(ctx context.Context, id uuid.UUID, checkpoint string) error
}

type DataSourceRepository struct {
	db *db.DB
}

func NewDataSourceRepository(db *db.DB) DataSourceRepositoryInt {
	return &DataSourceRepository{db: db}
}

const (
	CREATE_DATA_SOURCE = `
		INSERT INTO data_sources (
			organization_id, name, description, type, provider,
			status, events_today, total_events, last_event_at, last_sync_at,
			error_count, delayed_by_minutes, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, 0, 0, $7, $8, 0, 0, $9
		) RETURNING id, created_at, updated_at
	`
	UPDATE_DATA_SOURCE = `
		UPDATE data_sources SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			provider = COALESCE($6, provider),
			status = COALESCE($7, status),
			metadata = COALESCE($10, metadata),
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`
	GET_DATA_SOURCE   = `SELECT * FROM data_sources WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	LIST_DATA_SOURCES = `
		SELECT * FROM data_sources
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND ($2::varchar IS NULL OR status = $2)
			AND ($3::varchar IS NULL OR type = $3)
			AND ($4::varchar IS NULL OR name ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%')
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`
	LIST_ALL_ACTIVE_DATA_SOURCES = `
		SELECT * FROM data_sources
		WHERE deleted_at IS NULL AND status='active'
		ORDER BY created_at DESC
	`
	COUNT_DATA_SOURCES = `
		SELECT COUNT(*) FROM data_sources
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND ($2::varchar IS NULL OR status = $2)
			AND ($3::varchar IS NULL OR type = $3)
			AND ($4::varchar IS NULL OR name ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%')
	`
	UPDATE_HEALTH_METRICS = `
		UPDATE data_sources SET
			events_today = events_today + $1,
			total_events = total_events + $2,
			last_event_at = COALESCE($3, last_event_at),
			last_sync_at = $4,
			error_count = error_count + $5,
			updated_at = NOW()
		WHERE id = $6
	`
	GET_CHECKPOINT = `
	SELECT last_checkpoint
	FROM data_sources
	WHERE id = $1
	`
	UPDATE_CHECKPOINT = `
	UPDATE data_sources
	SET
		last_checkpoint = $2,
		last_checkpoint_at = NOW(),
		updated_at = NOW()
	WHERE id = $1
	`
	INCREMENT_EVENTS_TODAY = `UPDATE data_sources SET events_today = events_today + 1, updated_at = NOW() WHERE id = $1`
	RESET_DAILY_COUNTS     = `UPDATE data_sources SET events_today = 0, updated_at = NOW()`
	DISCONNECT_SOURCE      = `UPDATE data_sources SET status = $2, updated_at = NOW() WHERE id = $1 AND organization_id = $3`
	DELETE_SOURCE          = `UPDATE data_sources SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND organization_id = $2`
)

func (r *DataSourceRepository) CreateDataSource(ctx context.Context, ds *models.DataSource) error {
	var id uuid.UUID
	var createdAt, updatedAt time.Time
	metaJSON := ds.Metadata
	if metaJSON == nil {
		metaJSON = json.RawMessage{}
	}
	metaJSONMarshalled, err := json.Marshal(ds.Metadata)
	if err != nil {
		return err
	}
	err = r.db.QueryRowContext(
		ctx, CREATE_DATA_SOURCE,
		ds.OrganizationID, ds.Name, ds.Description, ds.Type, ds.Provider,
		ds.Status, ds.LastEventAt, ds.LastSyncAt, metaJSONMarshalled,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}
	ds.ID = id
	ds.CreatedAt = createdAt
	ds.UpdatedAt = updatedAt
	ds.Metadata = metaJSON
	return nil
}

func (r *DataSourceRepository) UpdateDataSource(ctx context.Context, ds *models.DataSource) error {
	metaJSON := ds.Metadata
	if metaJSON == nil {
		metaJSON = json.RawMessage{}
	}
	result, err := r.db.ExecContext(
		ctx, UPDATE_DATA_SOURCE,
		ds.ID, ds.OrganizationID, ds.Name, ds.Description, ds.Type, ds.Provider,
		ds.Status, ds.LastEventAt, ds.LastSyncAt, metaJSON,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("DATA SOURCE NOT FOUND")
	}
	ds.Metadata = metaJSON
	return nil
}

func (r *DataSourceRepository) GetDataSourceByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.DataSource, error) {
	var ds models.DataSource
	err := r.db.GetContext(ctx, &ds, GET_DATA_SOURCE, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("DATA SOURCE NOT FOUND")
		}
		return nil, err
	}
	return &ds, nil
}

func (r *DataSourceRepository) ListDataSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.DataSource, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	var status *string
	if s, ok := filters["status"].(string); ok && s != "" {
		status = &s
	}
	var typ *string
	if t, ok := filters["type"].(string); ok && t != "" {
		typ = &t
	}
	var search *string
	if s, ok := filters["search"].(string); ok && s != "" {
		search = &s
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_DATA_SOURCES, orgID, status, typ, search)
	if err != nil {
		return nil, 0, err
	}

	var sources []*models.DataSource
	err = r.db.SelectContext(ctx, &sources, LIST_DATA_SOURCES, orgID, status, typ, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return sources, total, nil
}

func (r *DataSourceRepository) UpdateHealthMetrics(ctx context.Context, id uuid.UUID, eventsToday, totalEvents, errorCount int64, lastEventAt, lastSyncAt *time.Time) error {
	result, err := r.db.ExecContext(ctx, UPDATE_HEALTH_METRICS,
		eventsToday, totalEvents, lastEventAt, lastSyncAt, errorCount, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("DATA SOURCE NOT FOUND")
	}
	return nil
}

func (r *DataSourceRepository) IncrementEventsToday(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, INCREMENT_EVENTS_TODAY, id)
	return err
}

func (r *DataSourceRepository) ResetDailyCounts(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, RESET_DAILY_COUNTS)
	return err
}

func (r *DataSourceRepository) DisconnectDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, DISCONNECT_SOURCE, id, models.DataSourceStatusDisconnected, orgID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("DATA SOURCE NOT FOUND")
	}
	return nil
}

func (r *DataSourceRepository) DeleteDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, DELETE_SOURCE, id, orgID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("DATA SOURCE NOT FOUND")
	}
	return nil
}

func (r *DataSourceRepository) GetAggregatedHealth(ctx context.Context, orgID uuid.UUID) (totalEvents, activeSources, delayedSources, errorSources int64, err error) {
	type healthResult struct {
		TotalEvents    int64 `db:"total_events"`
		ActiveSources  int64 `db:"active_sources"`
		DelayedSources int64 `db:"delayed_sources"`
		ErrorSources   int64 `db:"error_sources"`
	}
	var result healthResult
	query := `
		SELECT 
			COALESCE(SUM(total_events), 0) as total_events,
			COALESCE(SUM(CASE WHEN status = $2 THEN 1 ELSE 0 END), 0) as active_sources,
			COALESCE(SUM(CASE WHEN delayed_by_minutes > 0 THEN 1 ELSE 0 END), 0) as delayed_sources,
			COALESCE(SUM(CASE WHEN status = $3 THEN 1 ELSE 0 END), 0) as error_sources
		FROM data_sources
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	err = r.db.GetContext(ctx, &result, query, orgID, models.DataSourceStatusActive, models.DataSourceStatusError)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return result.TotalEvents, result.ActiveSources, result.DelayedSources, result.ErrorSources, nil
}

func (r *DataSourceRepository) GetSourcesWithIssues(ctx context.Context, orgID uuid.UUID) ([]*models.DataSource, error) {
	var sources []*models.DataSource
	err := r.db.SelectContext(ctx, &sources, `
		SELECT * FROM data_sources
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND (delayed_by_minutes > 15 OR status = $2)
		ORDER BY delayed_by_minutes DESC, error_count DESC
	`, orgID, models.DataSourceStatusError)
	if err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *DataSourceRepository) ListAllActiveDataSources(ctx context.Context) ([]*models.DataSource, error) {
	var sources []*models.DataSource
	err := r.db.SelectContext(ctx, &sources, LIST_ALL_ACTIVE_DATA_SOURCES)
	if err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *DataSourceRepository) GetCheckpoint(ctx context.Context, id uuid.UUID) (*string, error) {
	var checkpoint *string

	err := r.db.GetContext(
		ctx,
		&checkpoint,
		GET_CHECKPOINT,
		id,
	)

	if err != nil {
		return nil, err
	}

	return checkpoint, nil
}

func (r *DataSourceRepository) UpdateCheckpoint(ctx context.Context, id uuid.UUID, checkpoint string) error {

	_, err := r.db.ExecContext(
		ctx,
		UPDATE_CHECKPOINT,
		id,
		checkpoint,
	)

	return err
}
