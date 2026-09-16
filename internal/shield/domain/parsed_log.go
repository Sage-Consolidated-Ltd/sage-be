package domain

import (
	"database/sql"
	"sage-backend/internal/shared/db"
	"time"

	"github.com/google/uuid"
)

type ParsedLog struct {
	ID           uuid.UUID
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

type EventSearchParams struct {
	OrganizationID uuid.UUID
	DataSourceID   *uuid.UUID
	Source         *string
	Level          *string
	Severity       *string
	EventType      *string
	IngestionType  *string
	FreeText       *string
	RawFilters     map[string]string
	From, To       *time.Time
	Limit          int
	Cursor         *time.Time
}

type EventSearchResult struct {
	Events     []*SecurityEvent
	NextCursor *time.Time
	Total      int
}

type QueryAST struct {
	Level        *string
	DataSourceID *string
	Source       *string
	EventType    *string
	Channel      *string
	RawFilters   map[string]string
	Phrases      []string
}
