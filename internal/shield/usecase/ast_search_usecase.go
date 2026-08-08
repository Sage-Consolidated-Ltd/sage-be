package usecase

import (
	"context"
	"fmt"
	"strings"

	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type ASTSearchService struct {
	repo outbound.ParsedLogRepository
}

func NewASTSearchService(repo outbound.ParsedLogRepository) *ASTSearchService {
	return &ASTSearchService{repo: repo}
}

func (s *ASTSearchService) SearchLogsAST(ctx context.Context, orgID uuid.UUID, queryString string, limit int) (domain.SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	ast := s.ParseQueryString(queryString)
	params, err := s.ASTToSearchParams(ast)
	if err != nil {
		return domain.SearchResult{}, apperrors.BadException(err.Error())
	}
	params.OrganizationID = orgID
	params.Limit = limit

	return s.repo.Search(ctx, params)
}

func (s *ASTSearchService) ASTToSearchParams(ast domain.QueryAST) (domain.SearchParams, error) {
	params := domain.SearchParams{
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

func (s *ASTSearchService) ParseQueryString(input string) domain.QueryAST {
	ast := domain.QueryAST{
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
