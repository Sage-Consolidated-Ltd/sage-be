package domain

import (
	"database/sql"
	"sage-backend/internal/shared/db"
	"time"

	"github.com/google/uuid"
)

type ParsedLog struct {
	ID           uuid.UUID    `db:"id"`
	LogFileID    uuid.UUID    `db:"log_file_id"`
	DataSourceID uuid.UUID    `db:"data_source_id"`
	FileID       uuid.UUID    `db:"file_id"`
	Timestamp    sql.NullTime `db:"timestamp"`
	Level        string       `db:"level"`
	Message      string       `db:"message"`
	RawJSON      db.JSONMap   `db:"raw_json"`
}

type SearchParams struct {
	DataSourceID   *uuid.UUID
	OrganizationID uuid.UUID
	Level          *string
	FreeText       *string
	RawFilters     map[string]string
	From, To       *time.Time
	Limit          int
	Cursor         *time.Time
}

type SearchResult struct {
	Logs       []ParsedLog
	NextCursor *time.Time
}

type QueryAST struct {
	Level        *string
	DataSourceID *string
	RawFilters   map[string]string
	Phrases      []string
}
