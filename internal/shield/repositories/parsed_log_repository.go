package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/models"
)

const (
	BULK_STORE_PARSED_LOG = `
		INSERT INTO parsed_logs (id, data_source_id, file_id, timestamp, level, message, raw_json)
		SELECT * FROM unnest(
			$1::uuid[],
			$2::uuid[],
			$3::uuid[],
			$4::timestamptz[],
			$5::text[],
			$6::text[],
			$7::jsonb[]
		)
	`

	DELETE_PARSED_LOGS_BY_FILE_ID = `
		DELETE FROM parsed_logs WHERE file_id = $1
	`
)

type ParsedLogRepositoryInt interface {
	StoreParsedLogs(ctx context.Context, logs []models.ParsedLog) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	ReplaceParsedLogs(ctx context.Context, fileID uuid.UUID, logs []models.ParsedLog) error
}

type LogSearcher interface {
	Search(ctx context.Context, params models.SearchParams) (models.SearchResult, error)
}

type ParsedLogRepository struct {
	db *db.DB
}

func NewParsedLogRepository(db *db.DB) *ParsedLogRepository {
	return &ParsedLogRepository{
		db: db,
	}
}

func buildBulkInsertParsedLogsQuery(logs []models.ParsedLog) (string, []any, error) {
	const colsPerRow = 7
	valuePlaceholders := make([]string, len(logs))
	args := make([]any, 0, len(logs)*colsPerRow)

	for i, l := range logs {
		id := l.ID
		if id == uuid.Nil {
			id = uuid.New()
		}

		var ts any
		if !l.Timestamp.Time.IsZero() {
			ts = l.Timestamp
		}

		base := i * colsPerRow
		valuePlaceholders[i] = fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7,
		)
		args = append(args, id, l.DataSourceID, l.FileID, ts, l.Level, l.Message, l.RawJSON)
	}

	query := fmt.Sprintf(`
		INSERT INTO parsed_logs (id, data_source_id, file_id, timestamp, level, message, raw_json)
		VALUES %s
	`, strings.Join(valuePlaceholders, ","))

	return query, args, nil
}

func (p *ParsedLogRepository) StoreParsedLogs(ctx context.Context, logs []models.ParsedLog) error {
	if len(logs) == 0 {
		return nil
	}

	query, args, err := buildBulkInsertParsedLogsQuery(logs)
	if err != nil {
		return err
	}

	if _, err := p.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("bulk insert parsed logs: %w", err)
	}
	return nil
}

func (p *ParsedLogRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	if _, err := p.db.ExecContext(ctx, DELETE_PARSED_LOGS_BY_FILE_ID, fileID); err != nil {
		return fmt.Errorf("delete parsed logs for file %s: %w", fileID, err)
	}
	return nil
}

func (p *ParsedLogRepository) ReplaceParsedLogs(ctx context.Context, fileID uuid.UUID, logs []models.ParsedLog) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, DELETE_PARSED_LOGS_BY_FILE_ID, fileID); err != nil {
		return fmt.Errorf("delete existing parsed logs: %w", err)
	}

	if len(logs) > 0 {
		query, args, err := buildBulkInsertParsedLogsQuery(logs)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("bulk insert parsed logs: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *ParsedLogRepository) Search(ctx context.Context, params models.SearchParams) (models.SearchResult, error) {
	var logs []models.ParsedLog
	query, args, err := buildSearchQuery(params)
	if err != nil {
		return models.SearchResult{}, fmt.Errorf("build search query: %w", err)
	}

	if err := r.db.SelectContext(ctx, &logs, query, args...); err != nil {
		return models.SearchResult{}, fmt.Errorf("search parsed logs: %w", err)
	}

	var next *time.Time
	if len(logs) > 0 {
		var validLast *time.Time
		last := logs[len(logs)-1].Timestamp
		if last.Valid {
			validLast = &last.Time
		}
		next = validLast
	}

	return models.SearchResult{Logs: logs, NextCursor: next}, nil
}
func buildSearchQuery(params models.SearchParams) (string, []any, error) {
	if params.OrganizationID == uuid.Nil {
		return "", nil, fmt.Errorf("organization ID is required")
	}

	tableAlias := "pl"

	conds := []string{"1=1"}
	args := []any{params.OrganizationID}
	i := 2

	if params.DataSourceID != nil {
		conds = append(conds, fmt.Sprintf("%s.data_source_id = $%d", tableAlias, i))
		args = append(args, *params.DataSourceID)
		i++
	}
	if params.Level != nil {
		conds = append(conds, fmt.Sprintf("%s.level = $%d", tableAlias, i))
		args = append(args, *params.Level)
		i++
	}
	if params.From != nil {
		conds = append(conds, fmt.Sprintf("%s.timestamp >= $%d", tableAlias, i))
		args = append(args, *params.From)
		i++
	}
	if params.To != nil {
		conds = append(conds, fmt.Sprintf("%s.timestamp <= $%d", tableAlias, i))
		args = append(args, *params.To)
		i++
	}
	if params.FreeText != nil {
		conds = append(conds, fmt.Sprintf("%s.search_vector @@ plainto_tsquery('english', $%d)", tableAlias, i))
		args = append(args, *params.FreeText)
		i++
	}
	for k, v := range params.RawFilters {
		conds = append(conds, fmt.Sprintf("%s.raw_json ->> $%d::text = $%d", tableAlias, i, i+1))
		args = append(args, k, v)
		i += 2
	}
	if params.Cursor != nil {
		conds = append(conds, fmt.Sprintf("%s.timestamp < $%d", tableAlias, i))
		args = append(args, *params.Cursor)
		i++
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT pl.id, pl.data_source_id, pl.file_id, pl.timestamp, pl.level, pl.message, pl.raw_json
		FROM parsed_logs pl
		JOIN data_sources ds ON pl.data_source_id = ds.id
		WHERE %s
		AND ds.organization_id = $1
		ORDER BY pl.timestamp DESC
		LIMIT %d`, strings.Join(conds, " AND "), limit)

	return query, args, nil
}
