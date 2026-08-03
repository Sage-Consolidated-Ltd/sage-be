package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/usecase/dto"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockParserService struct {
	mock.Mock
}

func (m *mockParserService) GetParserSummary(ctx context.Context, orgID uuid.UUID) (int, int, float64, *time.Time, error) {
	args := m.Called(ctx, orgID)
	return args.Int(0), args.Int(1), args.Get(2).(float64), args.Get(3).(*time.Time), args.Error(4)
}

func (m *mockParserService) ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Parser, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.Parser), args.Get(1).(int), args.Error(2)
}

func (m *mockParserService) GetParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Parser, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Parser), args.Error(1)
}

func (m *mockParserService) CreateParser(ctx context.Context, parser *domain.Parser) error {
	args := m.Called(ctx, parser)
	return args.Error(0)
}

func (m *mockParserService) UpdateParser(ctx context.Context, parser *domain.Parser, changeNote string, changedBy uuid.UUID) error {
	args := m.Called(ctx, parser, changeNote, changedBy)
	return args.Error(0)
}

func (m *mockParserService) TestParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, sampleLog string, rawPayload map[string]interface{}) (*domain.ParserTestResponse, error) {
	args := m.Called(ctx, parserID, orgID, sampleLog, rawPayload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ParserTestResponse), args.Error(1)
}

func (m *mockParserService) PreviewParser(ctx context.Context, parserType types.ParserType, logic map[string]interface{}, mappings []map[string]interface{}, sampleLog string, rawPayload map[string]interface{}) (*domain.ParserTestResponse, error) {
	args := m.Called(ctx, parserType, logic, mappings, sampleLog, rawPayload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ParserTestResponse), args.Error(1)
}

func (m *mockParserService) EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *mockParserService) DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *mockParserService) ValidateParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockParserService) ValidateAllParsers(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockParserService) ImportParser(ctx context.Context, parser *domain.Parser) error {
	args := m.Called(ctx, parser)
	return args.Error(0)
}

func (m *mockParserService) ExportParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Parser, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Parser), args.Error(1)
}

func (m *mockParserService) ListSampleLogs(ctx context.Context, sourceID, parserID *uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	args := m.Called(ctx, sourceID, parserID, orgID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Get(1).(int), args.Error(2)
}

func TestNewParserHandler(t *testing.T) {
	handler := &ParserHandler{}
	assert.NotNil(t, handler)
}

func TestParserHandler_Struct(t *testing.T) {
	handler := &ParserHandler{service: nil}
	assert.NotNil(t, handler)
}

func TestParserHandler_GetParserSummary_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	now := time.Now()
	mockService.On("GetParserSummary", mock.Anything, orgID).Return(10, 8, 0.05, &now, nil)

	app.Get("/logs-data/parsers/summary", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetParserSummary(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers/summary", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_ListParsers_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parsers := []*domain.Parser{
		{ID: uuid.New(), Name: "AWS Parser", ParserType: "json"},
	}

	mockService.On("ListParsers", mock.Anything, orgID, mock.AnythingOfType("map[string]interface {}"), 1, 25).Return(parsers, 1, nil)

	app.Get("/logs-data/parsers", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ListParsers(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_GetParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parserID := uuid.New()
	parser := &domain.Parser{
		ID:   parserID,
		Name: "AWS Parser",
	}

	mockService.On("GetParser", mock.Anything, parserID, orgID).Return(parser, nil)

	app.Get("/logs-data/parsers/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetParser(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers/"+parserID.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_GetParser_InvalidID(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()

	app.Get("/logs-data/parsers/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetParser(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers/invalid-uuid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestParserHandler_CreateParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	req := dto.CreateParserRequest{
		Name:       "Test Parser",
		ParserType: "json",
	}

	mockService.On("CreateParser", mock.Anything, mock.AnythingOfType("*domain.Parser")).Return(nil)

	app.Post("/logs-data/parsers", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.CreateParser(c)
	})

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/logs-data/parsers", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(request)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_CreateParser_InvalidRequest(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()

	app.Post("/logs-data/parsers", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.CreateParser(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/parsers", strings.NewReader("invalid json"))
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestParserHandler_EnableParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parserID := uuid.New()

	mockService.On("EnableParser", mock.Anything, parserID, orgID).Return(nil)

	app.Post("/logs-data/parsers/:id/enable", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.EnableParser(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/parsers/"+parserID.String()+"/enable", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_DisableParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parserID := uuid.New()

	mockService.On("DisableParser", mock.Anything, parserID, orgID).Return(nil)

	app.Post("/logs-data/parsers/:id/disable", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DisableParser(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/parsers/"+parserID.String()+"/disable", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_ValidateParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parserID := uuid.New()
	result := map[string]interface{}{
		"job_id": uuid.New().String(),
	}

	mockService.On("ValidateParser", mock.Anything, parserID, orgID).Return(result, nil)

	app.Post("/logs-data/parsers/:id/validate", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ValidateParser(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/parsers/"+parserID.String()+"/validate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_ValidateAllParsers_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	result := map[string]interface{}{
		"job_id": uuid.New().String(),
	}

	mockService.On("ValidateAllParsers", mock.Anything, orgID).Return(result, nil)

	app.Post("/logs-data/parsers/validate", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ValidateAllParsers(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/parsers/validate", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_ExportParser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	parserID := uuid.New()
	exported := &domain.Parser{
		ID:   parserID,
		Name: "Test Parser",
	}

	mockService.On("ExportParser", mock.Anything, parserID, orgID).Return(exported, nil)

	app.Get("/logs-data/parsers/:id/export", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ExportParser(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers/"+parserID.String()+"/export", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestParserHandler_ListSampleLogs_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockParserService)
	handler := NewParserHandler(mockService)

	orgID := uuid.New()
	events := []*domain.SecurityEvent{}

	mockService.On("ListSampleLogs", mock.Anything, (*uuid.UUID)(nil), (*uuid.UUID)(nil), orgID, 1, 25).Return(events, 0, nil)

	app.Get("/logs-data/parsers/sample-logs", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ListSampleLogs(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/parsers/sample-logs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}
