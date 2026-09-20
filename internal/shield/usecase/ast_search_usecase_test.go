package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockASTEventRepo struct {
	mock.Mock
	mockEventRepo
}

func (m *mockASTEventRepo) SearchAST(ctx context.Context, params domain.EventSearchParams) (domain.EventSearchResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return domain.EventSearchResult{}, args.Error(1)
	}
	return args.Get(0).(domain.EventSearchResult), args.Error(1)
}

func TestASTSearchService_ParseQueryString(t *testing.T) {
	svc := NewASTSearchService(nil)

	input := `error "unauthorized access" level=critical source=okta type=login channel=upload raw.ip=192.168.1.1`
	ast := svc.ParseQueryString(input)

	assert.Contains(t, ast.Phrases, "error")
	assert.Contains(t, ast.Phrases, "unauthorized access")
	assert.NotNil(t, ast.Level)
	assert.Equal(t, "critical", *ast.Level)
	assert.NotNil(t, ast.DataSourceID)
	assert.Equal(t, "okta", *ast.DataSourceID)
	assert.NotNil(t, ast.EventType)
	assert.Equal(t, "login", *ast.EventType)
	assert.NotNil(t, ast.Channel)
	assert.Equal(t, "upload", *ast.Channel)
	assert.Equal(t, "192.168.1.1", ast.RawFilters["ip"])
}

func TestASTSearchService_ASTToSearchParams_SeverityAndChannel(t *testing.T) {
	svc := NewASTSearchService(nil)

	srcUUID := uuid.New()
	srcStr := srcUUID.String()
	lvl := "error"
	ch := "polled"
	evType := "user.login"

	ast := domain.QueryAST{
		Level:        &lvl,
		DataSourceID: &srcStr,
		EventType:    &evType,
		Channel:      &ch,
		Phrases:      []string{"failed", "password"},
		RawFilters:   map[string]string{"user": "john"},
	}

	params, err := svc.ASTToSearchParams(ast)
	assert.NoError(t, err)
	assert.NotNil(t, params.Severity)
	assert.Equal(t, "high", *params.Severity)
	assert.NotNil(t, params.DataSourceID)
	assert.Equal(t, srcUUID, *params.DataSourceID)
	assert.NotNil(t, params.EventType)
	assert.Equal(t, "user.login", *params.EventType)
	assert.NotNil(t, params.IngestionType)
	assert.Equal(t, "polled", *params.IngestionType)
	assert.NotNil(t, params.FreeText)
	assert.Equal(t, "failed password", *params.FreeText)
	assert.Equal(t, "john", params.RawFilters["user"])
}

func TestASTSearchService_SearchLogsAST_DelegatesToRepo(t *testing.T) {
	mockRepo := new(mockASTEventRepo)
	svc := NewASTSearchService(mockRepo)

	orgID := uuid.New()
	query := `level=warn channel=upload "disk full"`
	limit := 25

	now := time.Now()
	expectedResult := domain.EventSearchResult{
		Events: []*domain.SecurityEvent{
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				OccurredAt:     now,
			},
		},
		Total: 1,
	}

	mockRepo.On("SearchAST", mock.Anything, mock.MatchedBy(func(p domain.EventSearchParams) bool {
		return p.OrganizationID == orgID &&
			p.Severity != nil && *p.Severity == "medium" &&
			p.IngestionType != nil && *p.IngestionType == "upload" &&
			p.FreeText != nil && *p.FreeText == "disk full" &&
			p.Limit == limit
	})).Return(expectedResult, nil)

	res, err := svc.SearchLogsAST(context.Background(), orgID, query, limit)
	assert.NoError(t, err)
	assert.Equal(t, 1, res.Total)
	assert.Len(t, res.Events, 1)

	mockRepo.AssertExpectations(t)
}

