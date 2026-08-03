package outbound

import (
	"context"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type SecurityEventRepository interface {
	CreateEvent(ctx context.Context, event *domain.SecurityEvent) error
	BulkCreateEvents(ctx context.Context, events []*domain.SecurityEvent) error
	GetEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.SecurityEvent, error)
	SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error)
	GetEventsBySource(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error)
	UpdateParseStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status types.ParseStatus, errors []map[string]interface{}, normalized map[string]interface{}) error
	GetEventsByParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, limit int) ([]*domain.SecurityEvent, error)
	GetEventVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error)
	GetEventCountInWindow(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time) (int64, error)
	BulkCreateEventsWithReturning(ctx context.Context, events []*domain.SecurityEvent) ([]uuid.UUID, error)
	BulkInsertRawEvents(ctx context.Context, orgID uuid.UUID, sourceID *uuid.UUID, events []domain.NormalizedEvent) ([]domain.CreateRawEventResponse, error)
	GetRawEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.RawEvent, error)
}

type DataQualityRepository interface {
	CreateScan(ctx context.Context, scan *domain.DataQualityScan) error
	UpdateScan(ctx context.Context, scan *domain.DataQualityScan) error
	GetScanByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataQualityScan, error)
	ListScans(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]*domain.DataQualityScan, error)
	CreateSourceMetric(ctx context.Context, metric *domain.DataQualitySourceMetric) error
	GetSourceMetricsByScan(ctx context.Context, scanID uuid.UUID, orgID uuid.UUID) ([]*domain.DataQualitySourceMetric, error)
	CreateSuggestion(ctx context.Context, suggestion *domain.DataQualitySuggestion) error
	UpdateSuggestionStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.SuggestionStatus) error
	GetSuggestions(ctx context.Context, orgID uuid.UUID, sourceID, parserID *uuid.UUID, status domain.SuggestionStatus) ([]*domain.DataQualitySuggestion, error)
}

type ParserRepository interface {
	CreateParser(ctx context.Context, parser *domain.Parser) error
	UpdateParser(ctx context.Context, parser *domain.Parser) error
	GetParserByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Parser, error)
	ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Parser, int, error)
	EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	CreateParserVersion(ctx context.Context, version *domain.ParserVersion) error
	GetParserVersions(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID) ([]*domain.ParserVersion, error)
	CreateTestRun(ctx context.Context, run *domain.ParserTestRun) error
	GetParserSummary(ctx context.Context, orgID uuid.UUID) (total, active int, errorRate float64, lastUpdated *time.Time, err error)
	ImportParser(ctx context.Context, parser *domain.Parser) error
	IncrementParserMetrics(ctx context.Context, parserID uuid.UUID, eventsParsed int64, errorRate float64) error
}

type DataSourceRepository interface {
	CreateDataSource(ctx context.Context, ds *domain.DataSource) error
	UpdateDataSource(ctx context.Context, ds *domain.DataSource) error
	GetDataSourceByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataSource, error)
	ListDataSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.DataSource, int, error)
	UpdateHealthMetrics(ctx context.Context, id uuid.UUID, eventsToday, totalEvents, errorCount int64, lastEventAt, lastSyncAt *time.Time) error
	IncrementEventsToday(ctx context.Context, id uuid.UUID) error
	ResetDailyCounts(ctx context.Context) error
	DisconnectDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DeleteDataSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	GetAggregatedHealth(ctx context.Context, orgID uuid.UUID) (totalEvents, activeSources, delayedSources, errorSources int64, err error)
	GetSourcesWithIssues(ctx context.Context, orgID uuid.UUID) ([]*domain.DataSource, error)
	ListAllActiveDataSources(ctx context.Context) ([]*domain.DataSource, error)
	GetCheckpoint(ctx context.Context, id uuid.UUID) (*string, error)
	UpdateCheckpoint(ctx context.Context, id uuid.UUID, checkpoint string) error
}

type IntegrationRepository interface {
	CreateCredential(ctx context.Context, c *domain.IntegrationCredentials) error
	CreateDataSourceWithCredentialsBulk(ctx context.Context, creds *[]domain.IntegrationCredentials, ds *domain.DataSource) error
	GetCredentialsByIntegration(ctx context.Context, integrationID string) ([]domain.IntegrationCredentials, error)
}

type IngestionJobRepository interface {
	CreateJob(ctx context.Context, job *domain.IngestionJob) error
	GetJobByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.IngestionJob, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.JobStatus, eventsProcessed, eventsFailed int64, errMsg *string) error
	ListJobs(ctx context.Context, orgID uuid.UUID, jobType domain.JobType, status domain.JobStatus, page, pageSize int) ([]*domain.IngestionJob, int, error)
}

type LogUploadRepository interface {
	CreatePending(ctx context.Context, params domain.CreateLogFileParams) (*domain.LogFile, error)
	Confirm(ctx context.Context, params domain.ConfirmLogFileParams) (*domain.LogFile, error)
	GetByS3Key(ctx context.Context, s3Key string) (*domain.LogFile, error)
	MarkSubmitted(ctx context.Context, s3Key string) error
	MarkFailed(ctx context.Context, s3Key string, reason string) error
}

type ParsedLogRepository interface {
	StoreParsedLogs(ctx context.Context, logs []domain.ParsedLog) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	ReplaceParsedLogs(ctx context.Context, fileID uuid.UUID, logs []domain.ParsedLog) error
	Search(ctx context.Context, params domain.SearchParams) (domain.SearchResult, error)
}

type AnalysisRepository interface {
	RecordAnalysis(ctx context.Context, params *domain.CreateAnalysisParams) (*domain.AnalysisResult, error)
	GetByLogFileID(ctx context.Context, logFileID uuid.UUID) (*domain.AnalysisResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AnalysisResult, error)
	GetThreatsByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]domain.Threat, error)
}
