package models

import (
	"time"

	"github.com/google/uuid"
)

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
