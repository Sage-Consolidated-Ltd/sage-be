package dto

import (
	"time"

	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

// CreateParserRequest for creating a new parser
type CreateParserRequest struct {
	Name         string                   `json:"name" validate:"required,min=1,max=255"`
	Description  *string                  `json:"description,omitempty"`
	DataSourceID *uuid.UUID               `json:"data_source_id,omitempty"`
	ParserType   types.ParserType         `json:"parser_type" validate:"required,oneof=regex json csv key_value ai_nlp"`
	Tags         []string                 `json:"tags,omitempty"`
	Logic        map[string]interface{}   `json:"logic" validate:"required"`
	Mappings     []map[string]interface{} `json:"mappings,omitempty"`
	AutoGenerate bool                     `json:"auto_generate,omitempty"`
}

// UpdateParserRequest for updating existing parser
type UpdateParserRequest struct {
	Name         *string                  `json:"name,omitempty"`
	Description  *string                  `json:"description,omitempty"`
	DataSourceID *uuid.UUID               `json:"data_source_id,omitempty"`
	ParserType   *types.ParserType        `json:"parser_type,omitempty,oneof=regex json csv key_value ai_nlp"`
	Status       *types.ParserStatus      `json:"status,omitempty,oneof=active warning error disabled"`
	Tags         []string                 `json:"tags,omitempty"`
	Logic        map[string]interface{}   `json:"logic,omitempty"`
	Mappings     []map[string]interface{} `json:"mappings,omitempty"`
}

// TestParserRequest for testing a parser against sample
type TestParserRequest struct {
	SampleLog  string                 `json:"sample_log,omitempty"`
	RawPayload map[string]interface{} `json:"raw_payload,omitempty"`
}

// PreviewParserRequest for previewing parser output before creating
type PreviewParserRequest struct {
	ParserType types.ParserType         `json:"parser_type" validate:"required"`
	Logic      map[string]interface{}   `json:"logic" validate:"required"`
	Mappings   []map[string]interface{} `json:"mappings,omitempty"`
	SampleLog  string                   `json:"sample_log,omitempty"`
	RawPayload map[string]interface{}   `json:"raw_payload,omitempty"`
}

// ImportParserRequest for importing parser definition
type ImportParserRequest struct {
	Name         string                   `json:"name" validate:"required"`
	Description  *string                  `json:"description,omitempty"`
	DataSourceID *uuid.UUID               `json:"data_source_id,omitempty"`
	ParserType   types.ParserType         `json:"parser_type" validate:"required"`
	Status       types.ParserStatus       `json:"status"`
	Tags         []string                 `json:"tags,omitempty"`
	Logic        map[string]interface{}   `json:"logic"`
	Mappings     []map[string]interface{} `json:"mappings"`
}

// ApplySuggestedFixRequest for applying AI suggestion
type ApplySuggestedFixRequest struct {
	SuggestionID string  `json:"suggestion_id" validate:"required,uuid"`
	SourceID     *string `json:"source_id,omitempty"`
	ParserID     *string `json:"parser_id,omitempty"`
}

type ParserResponse struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Description     *string                  `json:"description,omitempty"`
	DataSource      *DataSourceResponse      `json:"data_source,omitempty"`
	ParserType      string                   `json:"parser_type"`
	Status          string                   `json:"status"`
	Tags            []string                 `json:"tags"`
	Logic           map[string]interface{}   `json:"logic"`
	Mappings        []map[string]interface{} `json:"mappings"`
	EventsParsed24h int64                    `json:"events_parsed_24h"`
	ErrorRate       float64                  `json:"error_rate"`
	Owner           *string                  `json:"owner,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}
