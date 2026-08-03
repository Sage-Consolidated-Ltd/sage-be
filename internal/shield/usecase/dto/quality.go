package dto

// GetSuggestedFixDiffQuery params
type GetSuggestedFixDiffRequest struct {
	SuggestionID string `json:"suggestion_id" query:"suggestion_id" validate:"required,uuid"`
	ParserID     string `json:"parser_id" query:"parser_id" validate:"required,uuid"`
}
