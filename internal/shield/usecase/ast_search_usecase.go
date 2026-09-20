package usecase

import (
	"context"
	"strings"

	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type ASTSearchService struct {
	repo outbound.SecurityEventRepository
}

func NewASTSearchService(repo outbound.SecurityEventRepository) *ASTSearchService {
	return &ASTSearchService{repo: repo}
}

func (s *ASTSearchService) SearchLogsAST(ctx context.Context, orgID uuid.UUID, queryString string, limit int) (domain.EventSearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	ast := s.ParseQueryString(queryString)
	params, err := s.ASTToSearchParams(ast)
	if err != nil {
		return domain.EventSearchResult{}, apperrors.BadException(err.Error())
	}
	params.OrganizationID = orgID
	params.Limit = limit

	return s.repo.SearchAST(ctx, params)
}

func (s *ASTSearchService) ASTToSearchParams(ast domain.QueryAST) (domain.EventSearchParams, error) {
	params := domain.EventSearchParams{
		RawFilters: ast.RawFilters,
	}

	if ast.Level != nil {
		lvl := strings.ToLower(strings.TrimSpace(*ast.Level))
		switch lvl {
		case "critical", "fatal":
			sev := "critical"
			params.Severity = &sev
		case "error", "err":
			sev := "high"
			params.Severity = &sev
		case "warn", "warning":
			sev := "medium"
			params.Severity = &sev
		case "info", "debug", "trace":
			sev := "low"
			params.Severity = &sev
		default:
			params.Severity = ast.Level
		}
		params.Level = ast.Level
	}

	if ast.DataSourceID != nil {
		id, err := uuid.Parse(*ast.DataSourceID)
		if err == nil {
			params.DataSourceID = &id
		} else {
			// If not a UUID, treat as source name
			params.Source = ast.DataSourceID
		}
	} else if ast.Source != nil {
		params.Source = ast.Source
	}

	if ast.EventType != nil {
		params.EventType = ast.EventType
	}

	if ast.Channel != nil {
		params.IngestionType = ast.Channel
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
			case key == "level" || key == "severity":
				ast.Level = &value
			case key == "source" || key == "data_source":
				ast.DataSourceID = &value
			case key == "type" || key == "event_type":
				ast.EventType = &value
			case key == "channel" || key == "ingest" || key == "mode":
				ast.Channel = &value
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
