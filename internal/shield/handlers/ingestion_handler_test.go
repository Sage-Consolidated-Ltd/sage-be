package handlers

import (
	"context"
	"net/http/httptest"
	"sage-backend/internal/shield/models"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogsDataService struct {
	mock.Mock
}

func (m *mockLogsDataService) GetIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockLogsDataService) RefreshIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockLogsDataService) ListSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.DataSource, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*models.DataSource), args.Get(1).(int), args.Error(2)
}

func (m *mockLogsDataService) GetSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.DataSource, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DataSource), args.Error(1)
}

func (m *mockLogsDataService) SyncSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockLogsDataService) DisconnectSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *mockLogsDataService) GetSourceLogs(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.SecurityEvent, int, error) {
	args := m.Called(ctx, sourceID, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*models.SecurityEvent), args.Get(1).(int), args.Error(2)
}

func (m *mockLogsDataService) GetIngestionVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error) {
	args := m.Called(ctx, orgID, startTime, endTime, interval, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockLogsDataService) GetIngestionNotifications(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockLogsDataService) DownloadIngestionHealthReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error) {
	args := m.Called(ctx, orgID, format, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Get(1).(string), args.Error(2)
	}
	return args.Get(0).([]byte), args.Get(1).(string), args.Error(2)
}

func TestNewLogsDataHandler(t *testing.T) {
	handler := &LogsDataHandler{}
	assert.NotNil(t, handler)
}

func TestLogsDataHandler_Struct(t *testing.T) {
	handler := &LogsDataHandler{service: nil}
	assert.NotNil(t, handler)
}

func TestLogsDataHandler_GetIngestionHealth_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	health := map[string]interface{}{
		"total_events":    1000,
		"active_sources":  5,
		"delayed_sources": 1,
		"error_sources":   0,
	}

	mockService.On("GetIngestionHealth", mock.Anything, orgID).Return(health, nil)

	app.Get("/logs-data/ingestion-health", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetIngestionHealth(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetIngestionHealth_Unauthorized(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	app.Get("/logs-data/ingestion-health", func(c *fiber.Ctx) error {
		return handler.GetIngestionHealth(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestLogsDataHandler_RefreshIngestionHealth_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	result := map[string]interface{}{
		"job_id": uuid.New().String(),
	}

	mockService.On("RefreshIngestionHealth", mock.Anything, orgID).Return(result, nil)

	app.Post("/logs-data/ingestion-health/refresh", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.RefreshIngestionHealth(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/ingestion-health/refresh", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_ListSources_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	sources := []*models.DataSource{
		{ID: uuid.New(), Name: "AWS CloudTrail", Type: "aws"},
	}

	mockService.On("ListSources", mock.Anything, orgID, mock.AnythingOfType("map[string]interface {}"), 1, 25).Return(sources, 1, nil)

	app.Get("/logs-data/sources", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.ListSources(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/sources", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetSource_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()
	source := &models.DataSource{
		ID:   sourceID,
		Name: "AWS CloudTrail",
		Type: "aws",
	}

	mockService.On("GetSource", mock.Anything, sourceID, orgID).Return(source, nil)

	app.Get("/logs-data/sources/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetSource(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/sources/"+sourceID.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetSource_InvalidID(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()

	app.Get("/logs-data/sources/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetSource(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/sources/invalid-uuid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLogsDataHandler_SyncSource_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()
	result := map[string]interface{}{
		"job_id": uuid.New().String(),
	}

	mockService.On("SyncSource", mock.Anything, sourceID, orgID).Return(result, nil)

	app.Post("/logs-data/sources/:id/sync", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.SyncSource(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/sources/"+sourceID.String()+"/sync", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_DisconnectSource_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()

	mockService.On("DisconnectSource", mock.Anything, sourceID, orgID).Return(nil)

	app.Post("/logs-data/sources/:id/disconnect", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DisconnectSource(c)
	})

	req := httptest.NewRequest("POST", "/logs-data/sources/"+sourceID.String()+"/disconnect", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetSourceLogs_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()
	logs := []*models.SecurityEvent{}

	mockService.On("GetSourceLogs", mock.Anything, sourceID, orgID, mock.AnythingOfType("map[string]interface {}"), 1, 25).Return(logs, 0, nil)

	app.Get("/logs-data/sources/:id/logs", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetSourceLogs(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/sources/"+sourceID.String()+"/logs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetIngestionVolume_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	volume := []map[string]interface{}{
		{"timestamp": time.Now().Format(time.RFC3339), "count": 100},
	}

	mockService.On("GetIngestionVolume", mock.Anything, orgID, (*time.Time)(nil), (*time.Time)(nil), "hour", (*uuid.UUID)(nil)).Return(volume, nil)

	app.Get("/logs-data/ingestion-health/volume", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetIngestionVolume(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health/volume", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_GetIngestionNotifications_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	notifs := map[string]interface{}{
		"warnings": []string{"Source delayed"},
	}

	mockService.On("GetIngestionNotifications", mock.Anything, orgID).Return(notifs, nil)

	app.Get("/logs-data/ingestion-health/notifications", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetIngestionNotifications(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health/notifications", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_DownloadIngestionHealthReport_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()
	data := []byte("csv,data")
	filename := "report.csv"

	mockService.On("DownloadIngestionHealthReport", mock.Anything, orgID, "csv", (*time.Time)(nil), (*time.Time)(nil)).Return(data, filename, nil)

	app.Get("/logs-data/ingestion-health/report", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DownloadIngestionHealthReport(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health/report?format=csv", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestLogsDataHandler_DownloadIngestionHealthReport_InvalidFormat(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsDataService)
	handler := NewLogsDataHandlerWithService(mockService)

	orgID := uuid.New()

	app.Get("/logs-data/ingestion-health/report", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.DownloadIngestionHealthReport(c)
	})

	req := httptest.NewRequest("GET", "/logs-data/ingestion-health/report?format=invalid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
