package domain

import (
	"sage-backend/internal/shared/types"
	"time"

	"github.com/google/uuid"
)

type Parser struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	SourceID        *uuid.UUID
	Name            string
	Description     *string
	ParserType      types.ParserType
	Status          types.ParserStatus
	Tags            []string
	Logic           map[string]interface{}
	Mappings        []map[string]interface{}
	EventsParsed24h int64
	ErrorRate       float64
	OwnerUserID     *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type ParserVersion struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ParserID       uuid.UUID
	VersionNumber  int
	Logic          map[string]interface{}
	Mappings       []map[string]interface{}
	ChangedBy      *uuid.UUID
	ChangeNote     *string
	CreatedAt      time.Time
}

type ParserTestRun struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	ParserID         *uuid.UUID
	SampleLog        *string
	RawPayload       map[string]interface{}
	ParsedOutput     map[string]interface{}
	NormalizedOutput map[string]interface{}
	Errors           []map[string]interface{}
	Success          bool
	CreatedAt        time.Time
}

type ParserTestResponse struct {
	Success          bool
	ParsedOutput     map[string]interface{}
	NormalizedOutput map[string]interface{}
	SchemaPreview    *map[string]interface{}
	Errors           []map[string]interface{}
}
