package models

import (
	"encoding/json"
	"sage-backend/internal/shared/types"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Parser struct {
	ID              uuid.UUID          `db:"id"`
	OrganizationID  uuid.UUID          `db:"organization_id"`
	Name            string             `db:"name"`
	Description     *string            `db:"description"`
	SourceID        *uuid.UUID         `db:"source_id"`
	ParserType      types.ParserType   `db:"parser_type"`
	Status          types.ParserStatus `db:"status"`
	Tags            pq.StringArray     `db:"tags"`
	Logic           json.RawMessage    `db:"logic"`
	Mappings        json.RawMessage    `db:"mappings"`
	EventsParsed24h int64              `db:"events_parsed_24h" json:"events_parsed_24h"`
	ErrorRate       float64            `db:"error_rate" json:"error_rate"`
	OwnerUserID     *uuid.UUID         `db:"owner_user_id,omitempty" json:"owner_user_id,omitempty"`
	CreatedAt       time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `db:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time         `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// type Parser struct {
// 	ID              uuid.UUID                `db:"id" json:"id"`
// 	OrganizationID  uuid.UUID                `db:"organization_id" json:"organization_id"`
// 	SourceID        *uuid.UUID               `db:"source_id,omitempty" json:"source_id,omitempty"`
// 	Name            string                   `db:"name" json:"name"`
// 	Description     *string                  `db:"description,omitempty" json:"description,omitempty"`
// 	ParserType      types.ParserType         `db:"parser_type" json:"parser_type"`
// 	Status          types.ParserStatus       `db:"status" json:"status"`
// 	Tags            []string                 `db:"tags" json:"tags"`
// 	Logic           map[string]interface{}   `db:"logic" json:"logic"`
// 	Mappings        []map[string]interface{} `db:"mappings" json:"mappings"`
// 	EventsParsed24h int64                    `db:"events_parsed_24h" json:"events_parsed_24h"`
// 	ErrorRate       float64                  `db:"error_rate" json:"error_rate"`
// 	OwnerUserID     *uuid.UUID               `db:"owner_user_id,omitempty" json:"owner_user_id,omitempty"`
// 	CreatedAt       time.Time                `db:"created_at" json:"created_at"`
// 	UpdatedAt       time.Time                `db:"updated_at" json:"updated_at"`
// 	DeletedAt       *time.Time               `db:"deleted_at,omitempty" json:"deleted_at,omitempty"`
// }

type ParserVersion struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	OrganizationID uuid.UUID       `db:"organization_id" json:"organization_id"`
	ParserID       uuid.UUID       `db:"parser_id" json:"parser_id"`
	VersionNumber  int             `db:"version_number" json:"version_number"`
	Logic          json.RawMessage `db:"logic" json:"logic"`
	Mappings       json.RawMessage `db:"mappings" json:"mappings"`
	ChangedBy      *uuid.UUID      `db:"changed_by,omitempty" json:"changed_by,omitempty"`
	ChangeNote     *string         `db:"change_note,omitempty" json:"change_note,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type ParserTestRun struct {
	ID               uuid.UUID                `db:"id" json:"id"`
	OrganizationID   uuid.UUID                `db:"organization_id" json:"organization_id"`
	ParserID         *uuid.UUID               `db:"parser_id,omitempty" json:"parser_id,omitempty"`
	SampleLog        *string                  `db:"sample_log,omitempty" json:"sample_log,omitempty"`
	RawPayload       map[string]interface{}   `db:"raw_payload,omitempty" json:"raw_payload,omitempty"`
	ParsedOutput     map[string]interface{}   `db:"parsed_output" json:"parsed_output"`
	NormalizedOutput map[string]interface{}   `db:"normalized_output" json:"normalized_output"`
	Errors           []map[string]interface{} `db:"errors" json:"errors"`
	Success          bool                     `db:"success" json:"success"`
	CreatedAt        time.Time                `db:"created_at" json:"created_at"`
}

type ParserResponse struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     *string             `json:"description,omitempty"`
	DataSource      *DataSourceResponse `json:"data_source,omitempty"`
	ParserType      string              `json:"parser_type"`
	Status          string              `json:"status"`
	Tags            []string            `json:"tags"`
	Logic           json.RawMessage     `json:"logic"`
	Mappings        json.RawMessage     `json:"mappings"`
	EventsParsed24h int64               `json:"events_parsed_24h"`
	ErrorRate       float64             `json:"error_rate"`
	Owner           *string             `json:"owner,omitempty"` // owner name
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

func (p *Parser) ToResponse(ownerName *string) *ParserResponse {
	var ds *DataSourceResponse
	if p.SourceID != nil {
		// To be filled by service if needed
		ds = &DataSourceResponse{ID: p.SourceID.String()}
	}
	return &ParserResponse{
		ID:              p.ID.String(),
		Name:            p.Name,
		Description:     p.Description,
		DataSource:      ds,
		ParserType:      string(p.ParserType),
		Status:          string(p.Status),
		Tags:            p.Tags,
		Logic:           p.Logic,
		Mappings:        p.Mappings,
		EventsParsed24h: p.EventsParsed24h,
		ErrorRate:       p.ErrorRate,
		Owner:           ownerName,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

type ParserTestResponse struct {
	Success          bool                     `json:"success"`
	ParsedOutput     map[string]interface{}   `json:"parsed_output"`
	NormalizedOutput map[string]interface{}   `json:"normalized_output"`
	Errors           []map[string]interface{} `json:"errors"`
	FieldMappings    []map[string]interface{} `json:"field_mappings,omitempty"`
	SchemaPreview    *map[string]interface{}  `json:"schema_preview,omitempty"`
}
