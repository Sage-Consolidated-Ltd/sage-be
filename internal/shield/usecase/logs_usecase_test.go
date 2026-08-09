package usecase

import (
	"context"
	"errors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEventRepo struct {
	mock.Mock
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, event *domain.SecurityEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepo) BulkCreateEvents(ctx context.Context, events []*domain.SecurityEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *mockEventRepo) GetEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.SecurityEvent, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityEvent), args.Error(1)
}

func (m *mockEventRepo) SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Get(1).(int), args.Error(2)
}

func (m *mockEventRepo) GetEventsBySource(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	args := m.Called(ctx, sourceID, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Get(1).(int), args.Error(2)
}

func (m *mockEventRepo) UpdateParseStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status types.ParseStatus, errors []map[string]interface{}, normalized map[string]interface{}) error {
	args := m.Called(ctx, id, orgID, status, errors, normalized)
	return args.Error(0)
}

func (m *mockEventRepo) GetEventsByParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, limit int) ([]*domain.SecurityEvent, error) {
	args := m.Called(ctx, parserID, orgID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Error(1)
}

func (m *mockEventRepo) GetEventVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error) {
	args := m.Called(ctx, orgID, startTime, endTime, interval, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockEventRepo) GetEventCountInWindow(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time) (int64, error) {
	args := m.Called(ctx, orgID, startTime, endTime)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockEventRepo) BulkCreateEventsWithReturning(ctx context.Context, events []*domain.SecurityEvent) ([]uuid.UUID, error) {
	args := m.Called(ctx, events)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *mockEventRepo) GetRawEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.RawEvent, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepo) BulkInsertRawEvents(ctx context.Context, orgID uuid.UUID, sourceID *uuid.UUID, events []domain.NormalizedEvent) ([]domain.CreateRawEventResponse, error) {
	args := m.Called(ctx, events)
	return nil, args.Error(0)
}

func (m *mockEventRepo) GetThreatsSummary(ctx context.Context, orgID uuid.UUID) (*domain.ThreatsSummary, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ThreatsSummary), args.Error(1)
}

type mockDataSourceRepo struct {
	mock.Mock
}

func (m *mockDataSourceRepo) CreateDataSource(ctx context.Context, ds *domain.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *mockDataSourceRepo) GetDataSourceByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataSource, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DataSource), args.Error(1)
}

func (m *mockDataSourceRepo) IncrementEventsToday(ctx context.Context, sourceID uuid.UUID) error {
	args := m.Called(ctx, sourceID)
	return args.Error(0)
}

func (m *mockDataSourceRepo) UpdateHealthMetrics(ctx context.Context, sourceID uuid.UUID, eventsToday, totalEvents, errorCount int64, lastEventAt, lastSyncAt *time.Time) error {
	args := m.Called(ctx, sourceID, eventsToday, totalEvents, errorCount, lastEventAt, lastSyncAt)
	return args.Error(0)
}

func (m *mockDataSourceRepo) ListDataSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.DataSource, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.DataSource), args.Get(1).(int), args.Error(2)
}

func (m *mockDataSourceRepo) UpdateDataSource(ctx context.Context, ds *domain.DataSource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *mockDataSourceRepo) DeleteDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *mockDataSourceRepo) DisconnectDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *mockDataSourceRepo) ResetDailyCounts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDataSourceRepo) GetAggregatedHealth(ctx context.Context, orgID uuid.UUID) (totalEvents, activeSources, delayedSources, errorSources int64, err error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(int64), args.Get(3).(int64), args.Error(4)
}

func (m *mockDataSourceRepo) GetSourcesWithIssues(ctx context.Context, orgID uuid.UUID) ([]*domain.DataSource, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DataSource), args.Error(1)
}
func (m *mockDataSourceRepo) ListAllActiveDataSources(ctx context.Context) ([]*domain.DataSource, error) {
	return nil, nil
}

func (m *mockDataSourceRepo) GetCheckpoint(ctx context.Context, id uuid.UUID) (*string, error) {
	return nil, nil
}
func (m *mockDataSourceRepo) UpdateCheckpoint(ctx context.Context, id uuid.UUID, checkpoint string) error {
	return nil
}

type mockJobRepo struct {
	mock.Mock
}

func (m *mockJobRepo) CreateJob(ctx context.Context, job *domain.IngestionJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *mockJobRepo) GetJobByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.IngestionJob, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IngestionJob), args.Error(1)
}

func (m *mockJobRepo) UpdateJobStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.JobStatus, eventsProcessed, eventsFailed int64, errMsg *string) error {
	args := m.Called(ctx, id, orgID, status, eventsProcessed, eventsFailed, errMsg)
	return args.Error(0)
}

func (m *mockJobRepo) ListJobs(ctx context.Context, orgID uuid.UUID, jobType domain.JobType, status domain.JobStatus, page, pageSize int) ([]*domain.IngestionJob, int, error) {
	args := m.Called(ctx, orgID, jobType, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.IngestionJob), args.Get(1).(int), args.Error(2)
}

func TestLogsService_Struct(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)
	assert.NotNil(t, service)
}

func TestLogsService_IngestLog_Success(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	dataSource := &domain.DataSource{
		ID:             sourceID,
		OrganizationID: orgID,
		Name:           "Test Source",
		Type:           "aws",
		Status:         "active",
	}

	req := &dto.IngestLogRequest{
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

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(dataSource, nil)
	mockEventRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.SecurityEvent")).Return(nil)
	// Async calls - don't assert on these as they run in goroutines
	mockDataSourceRepo.On("IncrementEventsToday", ctx, sourceID).Return(nil).Maybe()
	mockDataSourceRepo.On("UpdateHealthMetrics", ctx, sourceID, int64(1), int64(1), int64(0), mock.Anything, mock.Anything).Return(nil).Maybe()

	event, err := service.IngestLog(ctx, orgID, req)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, orgID, event.OrganizationID)
	assert.Equal(t, sourceID, event.SourceID)
	assert.Equal(t, req.EventType, event.EventType)
	assert.Equal(t, &req.Severity, event.Severity)

	mockDataSourceRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_IngestLog_InvalidSourceID(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()

	req := &dto.IngestLogRequest{
		SourceID:      "invalid-uuid",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	_, err := service.IngestLog(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_SOURCE_ID")
}

func TestLogsService_IngestLog_SourceNotFound(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	req := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(nil, errors.New("source not found"))

	_, err := service.IngestLog(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source not found")

	mockDataSourceRepo.AssertExpectations(t)
}

func TestLogsService_IngestLog_RepositoryError(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	dataSource := &domain.DataSource{
		ID:             sourceID,
		OrganizationID: orgID,
		Name:           "Test Source",
		Type:           "aws",
		Status:         "active",
	}

	req := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(dataSource, nil)
	mockEventRepo.On("CreateEvent", ctx, mock.AnythingOfType("*domain.SecurityEvent")).Return(errors.New("database error"))

	_, err := service.IngestLog(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockDataSourceRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_BulkIngestLogs_Success(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	dataSource := &domain.DataSource{
		ID:             sourceID,
		OrganizationID: orgID,
		Name:           "Test Source",
		Type:           "aws",
		Status:         "active",
	}

	eventReq := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		SourceEventID: "EVT-123",
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	req := &dto.BulkIngestLogsRequest{
		SourceID: sourceID.String(),
		Events:   []*dto.IngestLogRequest{eventReq},
	}

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(dataSource, nil)
	mockEventRepo.On("BulkCreateEvents", ctx, mock.AnythingOfType("[]*domain.SecurityEvent")).Return(nil)
	// Async calls - don't assert on these as they run in goroutines
	mockDataSourceRepo.On("IncrementEventsToday", ctx, sourceID).Return(nil).Maybe()
	mockDataSourceRepo.On("UpdateHealthMetrics", ctx, sourceID, int64(1), int64(1), int64(0), mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := service.BulkIngestLogs(ctx, orgID, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result["ingested"])
	assert.Equal(t, sourceID.String(), result["source_id"])
	assert.Equal(t, orgID.String(), result["organization_id"])

	mockDataSourceRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_BulkIngestLogs_InvalidSourceID(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()

	req := &dto.BulkIngestLogsRequest{
		SourceID: "invalid-uuid",
		Events:   []*dto.IngestLogRequest{},
	}

	_, err := service.BulkIngestLogs(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_SOURCE_ID")
}

func TestLogsService_BulkIngestLogs_SourceNotFound(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	eventReq := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	req := &dto.BulkIngestLogsRequest{
		SourceID: sourceID.String(),
		Events:   []*dto.IngestLogRequest{eventReq},
	}

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(nil, errors.New("source not found"))

	_, err := service.BulkIngestLogs(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source not found")

	mockDataSourceRepo.AssertExpectations(t)
}

func TestLogsService_BulkIngestLogs_RepositoryError(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	sourceID := uuid.New()

	dataSource := &domain.DataSource{
		ID:             sourceID,
		OrganizationID: orgID,
		Name:           "Test Source",
		Type:           "aws",
		Status:         "active",
	}

	eventReq := &dto.IngestLogRequest{
		SourceID:      sourceID.String(),
		EventType:     "user_login",
		EventCategory: "authentication",
		Severity:      types.SeverityMedium,
		OccurredAt:    time.Now(),
		RawPayload:    map[string]interface{}{"action": "login"},
	}

	req := &dto.BulkIngestLogsRequest{
		SourceID: sourceID.String(),
		Events:   []*dto.IngestLogRequest{eventReq},
	}

	mockDataSourceRepo.On("GetDataSourceByID", ctx, sourceID, orgID).Return(dataSource, nil)
	mockEventRepo.On("BulkCreateEvents", ctx, mock.AnythingOfType("[]*domain.SecurityEvent")).Return(errors.New("database error"))

	_, err := service.BulkIngestLogs(ctx, orgID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockDataSourceRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_SearchLogs_Success(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	filters := map[string]interface{}{"severity": "high"}
	events := []*domain.SecurityEvent{}

	mockEventRepo.On("SearchEvents", ctx, orgID, filters, 1, 25).Return(events, 0, nil)

	result, total, err := service.SearchLogs(ctx, orgID, filters, 1, 25)
	assert.NoError(t, err)
	assert.Equal(t, events, result)
	assert.Equal(t, 0, total)

	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_SearchLogs_RepositoryError(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	filters := map[string]interface{}{"severity": "high"}

	mockEventRepo.On("SearchEvents", ctx, orgID, filters, 1, 25).Return(nil, 0, errors.New("database error"))

	_, _, err := service.SearchLogs(ctx, orgID, filters, 1, 25)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_GetLogByID_Success(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	eventID := uuid.New()
	severity := types.SeverityMedium

	event := &domain.SecurityEvent{
		ID:             eventID,
		OrganizationID: orgID,
		EventType:      "user_login",
		Severity:       &severity,
	}

	mockEventRepo.On("GetEventByID", ctx, eventID, orgID).Return(event, nil)

	result, err := service.GetLogByID(ctx, orgID, eventID)
	assert.NoError(t, err)
	assert.Equal(t, event, result)

	mockEventRepo.AssertExpectations(t)
}

func TestLogsService_GetLogByID_NotFound(t *testing.T) {
	mockEventRepo := new(mockEventRepo)
	mockDataSourceRepo := new(mockDataSourceRepo)
	mockJobRepo := new(mockJobRepo)
	service := NewLogsService(mockEventRepo, mockDataSourceRepo, mockJobRepo, nil)

	ctx := context.Background()
	orgID := uuid.New()
	eventID := uuid.New()

	mockEventRepo.On("GetEventByID", ctx, eventID, orgID).Return(nil, errors.New("event not found"))

	_, err := service.GetLogByID(ctx, orgID, eventID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event not found")

	mockEventRepo.AssertExpectations(t)
}
