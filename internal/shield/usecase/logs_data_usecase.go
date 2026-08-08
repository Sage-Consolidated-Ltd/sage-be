package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/shield/adapters/outbound/queue"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)



type LogsDataService struct {
	dataSourceRepo outbound.DataSourceRepository
	eventRepo      outbound.SecurityEventRepository
	jobRepo        outbound.IngestionJobRepository
	taskClient     *queue.TaskClient
}

func NewLogsDataService(
	dsRepo outbound.DataSourceRepository,
	eventRepo outbound.SecurityEventRepository,
	jobRepo outbound.IngestionJobRepository,
	taskClient *queue.TaskClient,
) inbound.LogsDataUseCase {
	return &LogsDataService{
		dataSourceRepo: dsRepo,
		eventRepo:      eventRepo,
		jobRepo:        jobRepo,
		taskClient:     taskClient,
	}
}

func (s *LogsDataService) GetIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	// Get aggregated stats from data source repo
	totalEvents, activeSources, delayedSources, errorSources, err := s.dataSourceRepo.GetAggregatedHealth(ctx, orgID)
	if err != nil {
		return nil, err
	}
	// Compute delta: events ingested in last 24h
	var deltaEvents int64 = 0
	// Use eventRepo to count events in last 24h
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	count, err := s.eventRepo.GetEventCountInWindow(ctx, orgID, &twentyFourHoursAgo, nil)
	if err == nil {
		deltaEvents = count
	}
	// For simplicity, other deltas are zero until we have historical tracking
	health := map[string]interface{}{
		"total_events_ingested":       totalEvents,
		"total_events_ingested_delta": deltaEvents,
		"active_data_sources":         activeSources,
		"active_data_sources_delta":   int64(0),
		"delayed_sources":             delayedSources,
		"delayed_sources_delta":       int64(0),
		"ingestion_errors":            errorSources,
		"ingestion_errors_delta":      int64(0),
		"generated_at":                time.Now(),
	}
	return health, nil
}

func (s *LogsDataService) RefreshIngestionHealth(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	job := &domain.IngestionJob{
		OrganizationID: orgID,
		Status:         domain.JobStatusQueued,
		JobType:        domain.JobTypeQualityScan,
		Metadata:       map[string]interface{}{"action": "refresh_ingestion_health"},
	}
	if err := s.jobRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"job_id":    job.ID.String(),
		"status":    "queued",
		"queued_at": job.CreatedAt,
	}, nil
}

func (s *LogsDataService) ListSources(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.DataSource, int, error) {
	return s.dataSourceRepo.ListDataSources(ctx, orgID, filters, page, pageSize)
}

func (s *LogsDataService) GetSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.DataSource, error) {
	return s.dataSourceRepo.GetDataSourceByID(ctx, id, orgID)
}

func (s *LogsDataService) SyncSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error) {
	// Verify source exists and belongs to org
	_, err := s.dataSourceRepo.GetDataSourceByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.taskClient.EnqueueProviderSync(ctx, orgID, id); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"source_id": id.String(),
		"status":    "queued",
	}, nil
}

func (s *LogsDataService) DisconnectSource(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return s.dataSourceRepo.DisconnectDataSource(ctx, id, orgID)
}

func (s *LogsDataService) GetSourceLogs(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	return s.eventRepo.GetEventsBySource(ctx, sourceID, orgID, filters, page, pageSize)
}

func (s *LogsDataService) GetIngestionVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error) {
	return s.eventRepo.GetEventVolume(ctx, orgID, startTime, endTime, interval, sourceID)
}

func (s *LogsDataService) GetIngestionNotifications(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	// Warnings: sources with issues
	sources, err := s.dataSourceRepo.GetSourcesWithIssues(ctx, orgID)
	if err != nil {
		return nil, err
	}
	warnings := make([]map[string]interface{}, 0, len(sources))
	for _, src := range sources {
		msg := ""
		if src.Status == domain.DataSourceStatusError {
			msg = fmt.Sprintf("Source %s has encountered errors", src.Name)
		} else if src.DelayedByMinutes > 0 {
			msg = fmt.Sprintf("Source %s is delayed by %d minutes", src.Name, src.DelayedByMinutes)
		}
		sev := "low"
		if src.Status == domain.DataSourceStatusError {
			sev = "critical"
		} else if src.DelayedByMinutes > 60 {
			sev = "high"
		} else if src.DelayedByMinutes > 30 {
			sev = "medium"
		}
		warnings = append(warnings, map[string]interface{}{
			"id":          src.ID.String(),
			"source_name": src.Name,
			"message":     msg,
			"severity":    sev,
			"tags":        []string{"source_health"},
			"created_at":  time.Now(),
		})
	}

	// AI insights: could be from data quality suggestions; fetch latest suggestions
	// For now placeholder
	insights := []map[string]interface{}{
		{
			"id":             "ai-001",
			"title":          "Parser performance degraded",
			"message":        "Parser error rate increased by 15% in the last hour",
			"recommendation": "Review parser rules and update mappings",
			"created_at":     time.Now(),
		},
	}
	return map[string]interface{}{
		"warnings":    warnings,
		"ai_insights": insights,
	}, nil
}

func (s *LogsDataService) DownloadIngestionHealthReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error) {
	// Get health data
	health, err := s.GetIngestionHealth(ctx, orgID)
	if err != nil {
		return nil, "", err
	}

	// Get volume data
	volume, err := s.GetIngestionVolume(ctx, orgID, startTime, endTime, "hour", nil)
	if err != nil {
		return nil, "", err
	}

	// Get sources
	sources, _, err := s.ListSources(ctx, orgID, map[string]interface{}{}, 1, 1000)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("ingestion-health-report-%s.%s", time.Now().Format("20060102-150405"), format)

	// Generate report based on format
	switch format {
	case "json":
		report := map[string]interface{}{
			"health":       health,
			"volume":       volume,
			"sources":      sources,
			"generated_at": time.Now(),
		}
		data, err := json.Marshal(report)
		if err != nil {
			return nil, "", err
		}
		return data, filename, nil
	case "csv":
		// Simple CSV generation
		var csv strings.Builder
		csv.WriteString("Source Name,Type,Status,Events Today,Error Count,Delayed By Minutes\n")
		for _, src := range sources {
			csv.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d,%d\n", src.Name, src.Type, src.Status, src.EventsToday, src.ErrorCount, src.DelayedByMinutes))
		}
		return []byte(csv.String()), filename, nil
	case "pdf":
		// Placeholder for PDF generation - would require a PDF library
		// For now, return JSON as fallback
		report := map[string]interface{}{
			"health":       health,
			"volume":       volume,
			"sources":      sources,
			"generated_at": time.Now(),
			"note":         "PDF generation not implemented, returning JSON instead",
		}
		data, err := json.Marshal(report)
		if err != nil {
			return nil, "", err
		}
		return data, strings.Replace(filename, ".pdf", ".json", 1), nil
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}
