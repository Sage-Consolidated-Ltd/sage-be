package requests

import (
	"encoding/json"
	"sage-backend/internal/shared/types"

	"github.com/google/uuid"
)

// CreateParserRequest for creating a new parser
type CreateParserRequest struct {
	Name         string           `json:"name" validate:"required"`
	Description  *string          `json:"description,omitempty"`
	DataSourceID *uuid.UUID       `json:"data_source_id" validate:"required"`
	ParserType   types.ParserType `json:"parser_type" validate:"required"`
	Tags         []string         `json:"tags,omitempty"`

	RegexLogic   *RegexLogic   `json:"regex_logic,omitempty"`
	JSONLogic    *JSONLogic    `json:"json_logic,omitempty"`
	CSVLogic     *CSVLogic     `json:"csv_logic,omitempty"`
	KeyValueLogic *KeyValueLogic `json:"key_value_logic,omitempty"`

	Mappings     []FieldMapping `json:"mappings,omitempty"`
	AutoGenerate bool           `json:"auto_generate,omitempty"`

	Status types.ParserStatus `json:"status,omitempty" validate:"omitempty,oneof=active warning error disabled"`
}

type RegexLogic struct {
	Pattern string `json:"pattern" validate:"required"`
}

type JSONLogic struct {
	Path string `json:"path"`
}

type CSVLogic struct {
	Delimiter string `json:"delimiter"`
	HasHeader bool   `json:"has_header"`
}

type KeyValueLogic struct {
	PairSeparator string `json:"pair_separator"`
	KeyValueSep   string `json:"key_value_separator"`
}

type FieldMapping struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	DataType    string `json:"data_type,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// UpdateParserRequest for updating existing parser
type UpdateParserRequest struct {
	Name         *string             `json:"name,omitempty"`
	Description  *string             `json:"description,omitempty"`
	DataSourceID *uuid.UUID          `json:"data_source_id,omitempty"`
	ParserType   *types.ParserType   `json:"parser_type,omitempty" validate:"omitempty,oneof=regex json csv key_value ai_nlp"`
	Status       *types.ParserStatus `json:"status,omitempty" validate:"omitempty,oneof=active warning error disabled"`
	Tags         *[]string           `json:"tags,omitempty"`

	Logic     json.RawMessage `json:"logic,omitempty"`
	Mappings  *[]FieldMapping `json:"mappings,omitempty"`
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
