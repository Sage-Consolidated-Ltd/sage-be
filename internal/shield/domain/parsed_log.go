package domain

import (
	"database/sql"
	"sage-backend/internal/shared/db"
	"time"

	"github.com/google/uuid"
)

type ParsedLog struct {
	ID           uuid.UUID    `db:"id"             json:"id"`
	DataSourceID uuid.UUID    `db:"data_source_id" json:"data_source_id"`
	FileID       uuid.UUID    `db:"file_id"        json:"file_id"`
	Timestamp    sql.NullTime `db:"timestamp"      json:"timestamp"`
	Level        string       `db:"level"          json:"level"`
	Message      string       `db:"message"        json:"message"`
	RawJSON      db.JSONMap   `db:"raw_json"       json:"raw_json"`
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
