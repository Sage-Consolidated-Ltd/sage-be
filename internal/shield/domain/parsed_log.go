package domain

import (
	"database/sql"
	"sage-backend/internal/shared/db"
	"time"

	"github.com/google/uuid"
)

type ParsedLog struct {
	ID           uuid.UUID
	LogFileID    uuid.UUID
	DataSourceID uuid.UUID
	FileID       uuid.UUID
	Timestamp    sql.NullTime
	Level        string
	Message      string
	RawJSON      db.JSONMap
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
