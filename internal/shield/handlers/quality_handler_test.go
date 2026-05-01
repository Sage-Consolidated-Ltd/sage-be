package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/requests"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDataQualityService struct {
	mock.Mock
}

func (m *mockDataQualityService) GetSummary(ctx context.Context, orgID uuid.UUID) (*models.DataQualityScan, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataQualityScan), args.Error(1)
}

func (m *mockDataQualityService) RunScan(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockDataQualityService) GetBreakdown(ctx context.Context, orgID uuid.UUID, sourceID *uuid.UUID, page, pageSize int) ([]*models.DataQualitySourceMetric, error) {
	args := m.Called(ctx, orgID, sourceID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DataQualitySourceMetric), args.Error(1)
}

func (m *mockDataQualityService) GetAIAnalysis(ctx context.Context, orgID uuid.UUID, sourceID *uuid.UUID) ([]map[string]interface{}, error) {
	args := m.Called(ctx, orgID, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockDataQualityService) ApplySuggestedFix(ctx context.Context, orgID uuid.UUID, suggestionID uuid.UUID) error {
	args := m.Called(ctx, orgID, suggestionID)
	return args.Error(0)
}

func (m *mockDataQualityService) GetSuggestedFixDiff(ctx context.Context, suggestionID uuid.UUID, parserID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, suggestionID, parserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockDataQualityService) DownloadDataQualityReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error) {
	args := m.Called(ctx, orgID, format, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Get(1).(string), args.Error(2)
	}
	return args.Get(0).([]byte), args.Get(1).(string), args.Error(2)
}

func TestNewQualityHandler(t *testing.T) {
	handler := &QualityHandler{}
	assert.NotNil(t, handler)
}

func TestQualityHandler_Struct(t *testing.T) {
	handler := &QualityHandler{service: nil}
	assert.NotNil(t, handler)
}

func TestQualityHandler_GetDataQualitySummary_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	score := 85
	scan := &models.DataQualityScan{
		ID:             uuid.New(),
		OrganizationID: orgID,
		QualityScore:   &score,
	}

	mockService.On("GetSummary", mock.Anything, orgID).Return(scan, nil)

	app.Get("/logs-data/data-quality", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetDataQualitySummary(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_RunDataQualityScan_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	result := map[string]interface{}{
		"job_id": uuid.New().String(),
	}

	mockService.On("RunScan", mock.Anything, orgID).Return(result, nil)

	app.Post("/logs-data/data-quality/scan", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.RunDataQualityScan(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/data-quality/scan", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_GetDataQualityBreakdown_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	metrics := []*models.DataQualitySourceMetric{
		{ID: uuid.New(), SourceID: uuid.New(), ParsingErrors: 10},
	}

	mockService.On("GetBreakdown", mock.Anything, orgID, (*uuid.UUID)(nil), 1, 25).Return(metrics, nil)

	app.Get("/logs-data/data-quality/breakdown", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetDataQualityBreakdown(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/breakdown", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_GetAIAnalysis_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	insights := []map[string]interface{}{
		{"type": "parsing_error", "count": 10},
	}

	mockService.On("GetAIAnalysis", mock.Anything, orgID, (*uuid.UUID)(nil)).Return(insights, nil)

	app.Get("/logs-data/data-quality/ai-analysis", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetAIAnalysis(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/ai-analysis", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_ApplySuggestedFix_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	suggestionID := uuid.New()
	req := requests.ApplySuggestedFixRequest{
		SuggestionID: suggestionID.String(),
	}

	mockService.On("ApplySuggestedFix", mock.Anything, orgID, suggestionID).Return(nil)

	app.Post("/logs-data/data-quality/apply-suggested-fix", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ApplySuggestedFix(c)
	})

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/logs-data/data-quality/apply-suggested-fix", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(request)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_ApplySuggestedFix_InvalidRequest(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()

	app.Post("/logs-data/data-quality/apply-suggested-fix", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ApplySuggestedFix(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/data-quality/apply-suggested-fix", strings.NewReader("invalid json"))
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestQualityHandler_GetSuggestedFixDiff_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	suggestionID := uuid.New()
	parserID := uuid.New()
	diff := map[string]interface{}{
		"before": "old code",
		"after":  "new code",
	}

	mockService.On("GetSuggestedFixDiff", mock.Anything, suggestionID, parserID).Return(diff, nil)

	app.Get("/logs-data/data-quality/diff", func(c *fiber.Ctx) error {
		return handler.GetSuggestedFixDiff(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/diff?suggestion_id="+suggestionID.String()+"&parser_id="+parserID.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_GetSuggestedFixDiff_MissingParameters(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	app.Get("/logs-data/data-quality/diff", func(c *fiber.Ctx) error {
		return handler.GetSuggestedFixDiff(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/diff", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestQualityHandler_DownloadDataQualityReport_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()
	data := []byte("csv,data")
	filename := "report.csv"

	mockService.On("DownloadDataQualityReport", mock.Anything, orgID, "csv", (*time.Time)(nil), (*time.Time)(nil)).Return(data, filename, nil)

	app.Get("/logs-data/data-quality/report", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DownloadDataQualityReport(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/report?format=csv", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestQualityHandler_DownloadDataQualityReport_InvalidFormat(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDataQualityService)
	handler := NewQualityHandler(mockService)

	orgID := uuid.New()

	app.Get("/logs-data/data-quality/report", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DownloadDataQualityReport(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/data-quality/report?format=invalid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
