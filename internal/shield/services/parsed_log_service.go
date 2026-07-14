package services

import (
	"context"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"
	"strings"

	"github.com/google/uuid"
)

type ParsedLogService struct {
	repo     repositories.ParsedLogRepositoryInt
	searcher repositories.LogSearcher
}

func NewParsedLogService(repo repositories.ParsedLogRepositoryInt, searcher repositories.LogSearcher) *ParsedLogService {
	return &ParsedLogService{repo: repo, searcher: searcher}
}

func (s *ParsedLogService) SearchLogs(ctx context.Context, params models.SearchParams) (models.SearchResult, error) {
	// Default limit
	if params.Limit <= 0 {
		params.Limit = 50
	}

	// Prevent excessively large requests
	if params.Limit > 200 {
		params.Limit = 200
	}

	// Optional: require at least one search filter
	if params.Level == nil &&
		params.FreeText == nil &&
		params.DataSourceID == nil &&
		params.From == nil &&
		params.To == nil {
		return models.SearchResult{}, apperrors.BadException("at least one search filter is required")
	}

	return s.searcher.Search(ctx, params)
}
func (s *ParsedLogService) ASTToSearchParams(ast models.QueryAST) (models.SearchParams, error) {
	params := models.SearchParams{
		RawFilters: ast.RawFilters,
	}

	if ast.Level != nil {
		params.Level = ast.Level
	}

	if ast.DataSourceID != nil {
		id, err := uuid.Parse(*ast.DataSourceID)
		if err != nil {
			return params, fmt.Errorf("invalid source id %q: %w", *ast.DataSourceID, err)
		}
		params.DataSourceID = &id
	}

	if len(ast.Phrases) > 0 {
		joined := strings.Join(ast.Phrases, " ")
		params.FreeText = &joined
	}

	return params, nil
}
func (s *ParsedLogService) ParseQueryString(input string) models.QueryAST {
	ast := models.QueryAST{
		RawFilters: make(map[string]string),
	}

	for _, tok := range tokenize(input) {
		switch {
		case strings.HasPrefix(tok, `"`) && strings.HasSuffix(tok, `"`) && len(tok) >= 2:
			phrase := strings.Trim(tok, `"`)
			if phrase != "" {
				ast.Phrases = append(ast.Phrases, phrase)
			}

		case strings.Contains(tok, "="):
			key, value, _ := strings.Cut(tok, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.Trim(strings.TrimSpace(value), `"`)

			switch {
			case key == "level":
				ast.Level = &value
			case key == "source":
				ast.DataSourceID = &value
			case strings.HasPrefix(key, "raw."):
				field := strings.TrimPrefix(key, "raw.")
				if field != "" {
					ast.RawFilters[field] = value
				}
			}

		default:
			if tok != "" {
				ast.Phrases = append(ast.Phrases, tok)
			}
		}
	}

	return ast
}
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, r := range input {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
