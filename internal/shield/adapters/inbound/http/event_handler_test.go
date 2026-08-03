package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/mocks"
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

type mockLogsService struct {
	mock.Mock
}

func (m *mockLogsService) IngestLog(ctx context.Context, orgID uuid.UUID, req *dto.IngestLogRequest) (*domain.SecurityEvent, error) {
	args := m.Called(ctx, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityEvent), args.Error(1)
}

func (m *mockLogsService) BulkIngestLogs(ctx context.Context, orgID uuid.UUID, req *dto.BulkIngestLogsRequest) (map[string]interface{}, error) {
	args := m.Called(ctx, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockLogsService) SearchLogs(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Get(1).(int), args.Error(2)
}

func (m *mockLogsService) GetLogByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*domain.SecurityEvent, error) {
	args := m.Called(ctx, orgID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityEvent), args.Error(1)
}

func TestNewEventHandler(t *testing.T) {
	mockService := new(mockLogsService)
	handler := &EventHandler{service: mockService}
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}

func TestEventHandler_GetLogDetail_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	event := mocks.GenerateMockSecurityEvent(orgID)

	mockService.On("GetLogByID", mock.Anything, mock.Anything, event.ID).Return(event, nil)

	app.Get("/logs/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetLogDetail(c)
	})

	req := httptest.NewRequest("GET", "/logs/"+event.ID.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Event retrieved", result["message"])
	assert.NotNil(t, result["data"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_GetLogDetail_InvalidID(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()

	app.Get("/logs/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetLogDetail(c)
	})

	req := httptest.NewRequest("GET", "/logs/invalid-uuid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestEventHandler_GetLogDetail_NotFound(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	eventID := uuid.New()

	mockService.On("GetLogByID", mock.Anything, orgID, eventID).Return(nil, assert.AnError)

	app.Get("/logs/:id", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetLogDetail(c)
	})

	req := httptest.NewRequest("GET", "/logs/"+eventID.String(), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestEventHandler_SearchLogs_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	events := mocks.GenerateMockSecurityEvents(orgID, 5)
	eventPtrs := make([]*domain.SecurityEvent, len(events))
	for i := range events {
		eventPtrs[i] = &events[i]
	}

	filters := map[string]interface{}{
		"source_id":      "",
		"source":         "",
		"event_type":     "",
		"event_category": "",
		"severity":       "",
		"actor_email":    "",
		"ip_address":     "",
		"start_time":     "",
		"end_time":       "",
		"search":         "",
	}

	mockService.On("SearchLogs", mock.Anything, orgID, filters, 1, 25).Return(eventPtrs, 5, nil)

	app.Get("/logs", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.SearchLogs(c)
	})

	req := httptest.NewRequest("GET", "/logs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Logs retrieved", result["message"])
	data := result["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(25), data["page_size"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_SearchLogs_WithFilters(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	events := mocks.GenerateMockSecurityEvents(orgID, 3)
	eventPtrs := make([]*domain.SecurityEvent, len(events))
	for i := range events {
		eventPtrs[i] = &events[i]
	}

	filters := map[string]interface{}{
		"source_id":      uuid.New().String(),
		"source":         "aws",
		"event_type":     "user_login",
		"event_category": "",
		"severity":       "high",
		"actor_email":    "test@example.com",
		"ip_address":     "192.168.1.1",
		"start_time":     "",
		"end_time":       "",
		"search":         "test",
	}

	mockService.On("SearchLogs", mock.Anything, orgID, filters, 1, 25).Return(eventPtrs, 3, nil)

	app.Get("/logs", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.SearchLogs(c)
	})

	req := httptest.NewRequest("GET", "/logs?source_id="+filters["source_id"].(string)+"&source=aws&event_type=user_login&severity=high&actor_email=test@example.com&ip_address=192.168.1.1&search=test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestEventHandler_SearchLogs_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()

	filters := map[string]interface{}{
		"source_id":      "",
		"source":         "",
		"event_type":     "",
		"event_category": "",
		"severity":       "",
		"actor_email":    "",
		"ip_address":     "",
		"start_time":     "",
		"end_time":       "",
		"search":         "",
	}

	mockService.On("SearchLogs", mock.Anything, mock.Anything, filters, 1, 25).Return(nil, 0, assert.AnError)

	app.Get("/logs", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.SearchLogs(c)
	})

	req := httptest.NewRequest("GET", "/logs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestEventHandler_IngestLog_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()
	event := mocks.GenerateMockSecurityEvent(orgID)

	reqBody := dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		SourceEventID: "EVT-123",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		ActorEmail:    "test@example.com",
		ActorUsername: "testuser",
		IPAddress:     "192.168.1.1",
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	mockService.On("IngestLog", mock.Anything, orgID, mock.MatchedBy(func(req *dto.IngestLogRequest) bool {
		return req.SourceID == reqBody.SourceID &&
			req.SourceEventID == reqBody.SourceEventID &&
			req.EventType == reqBody.EventType &&
			req.EventCategory == reqBody.EventCategory &&
			req.Severity == reqBody.Severity &&
			req.ActorEmail == reqBody.ActorEmail &&
			req.ActorUsername == reqBody.ActorUsername &&
			req.IPAddress == reqBody.IPAddress
	})).Return(event, nil)

	app.Post("/logs/ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.IngestLog(c)
	})

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/logs/ingest", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Event ingested", result["message"])
	assert.NotNil(t, result["data"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_IngestLog_InvalidRequest(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()

	app.Post("/logs/ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.IngestLog(c)
	})

	req := httptest.NewRequest("POST", "/logs/ingest", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestEventHandler_IngestLog_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()

	reqBody := dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		SourceEventID: "EVT-123",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	mockService.On("IngestLog", mock.Anything, orgID, mock.MatchedBy(func(req *dto.IngestLogRequest) bool {
		return req.SourceID == reqBody.SourceID &&
			req.SourceEventID == reqBody.SourceEventID &&
			req.EventType == reqBody.EventType &&
			req.EventCategory == reqBody.EventCategory &&
			req.Severity == reqBody.Severity
	})).Return(nil, assert.AnError)

	app.Post("/logs/ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.IngestLog(c)
	})

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/logs/ingest", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestEventHandler_BulkIngestLogs_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()

	eventReq := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		SourceEventID: "EVT-123",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	reqBody := dto.BulkIngestLogsRequest{
		SourceID: sourceID.String(),
		Events:   []*dto.IngestLogRequest{eventReq},
	}

	result := map[string]interface{}{
		"ingested":        1,
		"source_id":       sourceID.String(),
		"organization_id": orgID.String(),
	}

	mockService.On("BulkIngestLogs", mock.Anything, orgID, mock.MatchedBy(func(req *dto.BulkIngestLogsRequest) bool {
		return req.SourceID == reqBody.SourceID &&
			len(req.Events) == len(reqBody.Events)
	})).Return(result, nil)

	app.Post("/logs/bulk-ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.BulkIngestLogs(c)
	})

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/logs/bulk-ingest", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "Bulk ingestion completed", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestEventHandler_BulkIngestLogs_InvalidRequest(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()

	app.Post("/logs/bulk-ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.BulkIngestLogs(c)
	})

	req := httptest.NewRequest("POST", "/logs/bulk-ingest", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestEventHandler_BulkIngestLogs_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(mockLogsService)
	handler := NewEventHandler(mockService)

	orgID := uuid.New()
	sourceID := uuid.New()

	eventReq := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		SourceEventID: "EVT-123",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	reqBody := dto.BulkIngestLogsRequest{
		SourceID: sourceID.String(),
		Events:   []*dto.IngestLogRequest{eventReq},
	}

	mockService.On("BulkIngestLogs", mock.Anything, orgID, mock.MatchedBy(func(req *dto.BulkIngestLogsRequest) bool {
		return req.SourceID == reqBody.SourceID
	})).Return(nil, assert.AnError)

	app.Post("/logs/bulk-ingest", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.BulkIngestLogs(c)
	})

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/logs/bulk-ingest", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	mockService.AssertExpectations(t)
}
