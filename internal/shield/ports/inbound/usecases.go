package inbound

import (
	"context"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
	"time"

	"github.com/google/uuid"
)

type LogsUseCase interface {
	IngestLog(ctx context.Context, orgID uuid.UUID, req *dto.IngestLogRequest) (*domain.SecurityEvent, error)
	BulkIngestLogs(ctx context.Context, orgID uuid.UUID, req *dto.BulkIngestLogsRequest) (map[string]interface{}, error)
	SearchLogs(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error)
	SearchLogsAST(ctx context.Context, orgID uuid.UUID, queryString string, limit int) (domain.SearchResult, error)
	GetLogByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*domain.SecurityEvent, error)
}

type LogsDataUseCase interface {
	GetIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	RefreshIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	ListSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.DataSource, int, error)
	GetSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataSource, error)
	SyncSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error)
	DisconnectSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	GetSourceLogs(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error)
	GetIngestionVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error)
	GetIngestionNotifications(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	DownloadIngestionHealthReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error)
}

type DataQualityUseCase interface {
	GetSummary(ctx context.Context, orgID uuid.UUID) (*domain.DataQualityScan, error)
	RunScan(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	GetBreakdown(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID, page, pageSize int) ([]*domain.DataQualitySourceMetric, error)
	GetAIAnalysis(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID) ([]map[string]interface{}, error)
	ApplySuggestedFix(ctx context.Context, orgID uuid.UUID, suggestionID uuid.UUID) error
	GetSuggestedFixDiff(ctx context.Context, suggestionID, parserID uuid.UUID) (map[string]interface{}, error)
	DownloadDataQualityReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error)
}

type ParserUseCase interface {
	GetParserSummary(ctx context.Context, orgID uuid.UUID) (total, active int, errorRate float64, lastUpdated *time.Time, err error)
	ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Parser, int, error)
	GetParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Parser, error)
	CreateParser(ctx context.Context, parser *domain.Parser) error
	UpdateParser(ctx context.Context, parser *domain.Parser, changeNote string, changedBy uuid.UUID) error
	TestParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, sampleLog string, rawPayload map[string]interface{}) (*domain.ParserTestResponse, error)
	PreviewParser(ctx context.Context, parserType types.ParserType, logic map[string]interface{}, mappings []map[string]interface{}, sampleLog string, rawPayload map[string]interface{}) (*domain.ParserTestResponse, error)
	EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	ValidateParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error)
	ValidateAllParsers(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	ImportParser(ctx context.Context, parser *domain.Parser) error
	ExportParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Parser, error)
	ListSampleLogs(ctx context.Context, sourceID, parserID *uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]*domain.SecurityEvent, int, error)
}

type IntegrationUseCase interface {
	CreateDataSource(ctx context.Context, orgID uuid.UUID, req dto.CreateIntegrationRequest) error
}
