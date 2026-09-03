package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"
	"sage-backend/internal/shield/usecase"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockJobRepo struct {
	mock.Mock
	outbound.IngestionJobRepository
}

func (m *mockJobRepo) CreateJob(ctx context.Context, job *domain.IngestionJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *mockJobRepo) UpdateJobStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.JobStatus, processed, failed int64, errMsg *string) error {
	return m.Called(ctx, id, orgID, status, processed, failed, errMsg).Error(0)
}
func (m *mockJobRepo) GetJobByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.IngestionJob, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IngestionJob), args.Error(1)
}

type mockQualityRepo struct {
	mock.Mock
	outbound.DataQualityRepository
}

func (m *mockQualityRepo) CreateScan(ctx context.Context, scan *domain.DataQualityScan) error {
	return m.Called(ctx, scan).Error(0)
}
func (m *mockQualityRepo) UpdateScan(ctx context.Context, scan *domain.DataQualityScan) error {
	return m.Called(ctx, scan).Error(0)
}
func (m *mockQualityRepo) GetScanByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataQualityScan, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DataQualityScan), args.Error(1)
}
func (m *mockQualityRepo) CreateSourceMetric(ctx context.Context, metric *domain.DataQualitySourceMetric) error {
	return m.Called(ctx, metric).Error(0)
}

type mockEventRepo struct {
	mock.Mock
	outbound.SecurityEventRepository
}

func (m *mockEventRepo) SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	args := m.Called(ctx, orgID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.SecurityEvent), args.Int(1), args.Error(2)
}

func (m *mockEventRepo) BulkCreateEvents(ctx context.Context, events []*domain.SecurityEvent) error {
	return m.Called(ctx, events).Error(0)
}

func TestHandleQualityScanJob_EvaluatesEventsAndUpdatesMetrics(t *testing.T) {
	jobRepo := new(mockJobRepo)
	qualityRepo := new(mockQualityRepo)
	eventRepo := new(mockEventRepo)
	engine := usecase.NewDataQualityEngine()

	handler := NewTaskHandler(
		jobRepo,
		nil,
		eventRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		engine,
		qualityRepo,
	)

	orgID := uuid.New()
	jobID := uuid.New()
	scanID := uuid.New()
	sourceID := uuid.New()

	job := &domain.IngestionJob{
		ID:             jobID,
		OrganizationID: orgID,
		Status:         domain.JobStatusQueued,
	}
	scan := &domain.DataQualityScan{
		ID:             scanID,
		OrganizationID: orgID,
		Status:         "running",
	}

	eventID := "4624"
	severity := types.SeverityHigh
	testEvent := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		SourceID:       sourceID,
		SourceEventID:  &eventID,
		Source:         "SecurityLog",
		EventType:      "4624",
		Severity:       &severity,
		ParseStatus:    types.ParseStatusSuccess,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"channel":  "Security",
			"provider": "Microsoft-Windows-Security-Auditing",
			"user":     "alice",
		},
	}

	jobRepo.On("GetJobByID", mock.Anything, jobID, orgID).Return(job, nil)
	qualityRepo.On("GetScanByID", mock.Anything, scanID, orgID).Return(scan, nil)
	jobRepo.On("UpdateJobStatus", mock.Anything, jobID, orgID, domain.JobStatusRunning, int64(0), int64(0), (*string)(nil)).Return(nil)
	qualityRepo.On("UpdateScan", mock.Anything, mock.MatchedBy(func(s *domain.DataQualityScan) bool {
		return s.Status == "running"
	})).Return(nil)

	eventRepo.On("SearchEvents", mock.Anything, orgID, map[string]interface{}(nil), 1, 1000).Return([]*domain.SecurityEvent{testEvent}, 1, nil)
	qualityRepo.On("CreateSourceMetric", mock.Anything, mock.MatchedBy(func(m *domain.DataQualitySourceMetric) bool {
		return m.SourceID == sourceID && m.ScanID == scanID
	})).Return(nil)

	qualityRepo.On("UpdateScan", mock.Anything, mock.MatchedBy(func(s *domain.DataQualityScan) bool {
		return s.Status == "completed" && s.QualityScore != nil
	})).Return(nil)

	jobRepo.On("UpdateJobStatus", mock.Anything, jobID, orgID, domain.JobStatusCompleted, int64(1), int64(0), (*string)(nil)).Return(nil)

	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"job_id":          jobID,
		"organization_id": orgID,
		"scan_id":         scanID,
	})
	task := asynq.NewTask("data_quality:scan", payloadBytes)

	err := handler.HandleQualityScanJob(context.Background(), task)
	assert.NoError(t, err)

	jobRepo.AssertExpectations(t)
	qualityRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

