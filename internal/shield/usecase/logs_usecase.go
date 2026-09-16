package usecase

import (
	"context"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type LogsService struct {
	eventRepo      outbound.SecurityEventRepository
	dataSourceRepo outbound.DataSourceRepository
	jobRepo        outbound.IngestionJobRepository
	astSearcher    *ASTSearchService
}

func NewLogsService(
	eventRepo outbound.SecurityEventRepository,
	dataSourceRepo outbound.DataSourceRepository,
	jobRepo outbound.IngestionJobRepository,
	parsedLogRepo outbound.ParsedLogRepository,
) inbound.LogsUseCase {
	var astSearcher *ASTSearchService
	if eventRepo != nil {
		astSearcher = NewASTSearchService(eventRepo)
	}
	return &LogsService{
		eventRepo:      eventRepo,
		dataSourceRepo: dataSourceRepo,
		jobRepo:        jobRepo,
		astSearcher:    astSearcher,
	}
}

func (s *LogsService) SearchLogsAST(ctx context.Context, orgID uuid.UUID, queryString string, limit int) (domain.EventSearchResult, error) {
	if s.astSearcher == nil {
		return domain.EventSearchResult{}, nil
	}
	return s.astSearcher.SearchLogsAST(ctx, orgID, queryString, limit)
}

func (s *LogsService) IngestLog(ctx context.Context, orgID uuid.UUID, req *dto.IngestLogRequest) (*domain.SecurityEvent, error) {
	sourceUUID, err := uuid.Parse(req.SourceID)
	if err != nil {
		return nil, apperrors.BadException("INVALID_SOURCE_ID")
	}
	// Validate source belongs to org
	_, err = s.dataSourceRepo.GetDataSourceByID(ctx, sourceUUID, orgID)
	if err != nil {
		return nil, err
	}
	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	event := &domain.SecurityEvent{
		OrganizationID:    orgID,
		SourceID:          sourceUUID,
		SourceEventID:     &req.SourceEventID,
		Source:            "", // can be enriched later from data source
		EventType:         req.EventType,
		EventCategory:     types.EventCategory(req.EventCategory),
		Severity:          &req.Severity,
		ActorEmail:        &req.ActorEmail,
		ActorUsername:     &req.ActorUsername,
		IPAddress:         &req.IPAddress,
		GeoCountry:        &req.GeoCountry,
		GeoCity:           &req.GeoCity,
		RawPayload:        req.RawPayload,
		NormalizedPayload: db.JSONMap{"_ingest_type": "api_stream"},
		ParseStatus:       types.ParseStatusPending,
		ParseErrors:       make([]db.JSONMap, 0),
		OccurredAt:        occurredAt,
	}
	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return nil, err
	}
	// Update source metrics (async ideally)
	go func() {
		_ = s.dataSourceRepo.IncrementEventsToday(ctx, sourceUUID)
		_ = s.dataSourceRepo.UpdateHealthMetrics(ctx, sourceUUID, 1, 1, 0, nil, nil)
	}()
	return event, nil
}

func (s *LogsService) BulkIngestLogs(ctx context.Context, orgID uuid.UUID, req *dto.BulkIngestLogsRequest) (map[string]interface{}, error) {
	sourceUUID, err := uuid.Parse(req.SourceID)
	if err != nil {
		return nil, apperrors.BadException("INVALID_SOURCE_ID")
	}
	// Verify source belongs to org
	_, err = s.dataSourceRepo.GetDataSourceByID(ctx, sourceUUID, orgID)
	if err != nil {
		return nil, err
	}
	events := make([]*domain.SecurityEvent, 0, len(req.Events))
	for _, e := range req.Events {
		occurredAt := e.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = time.Now()
		}
		event := &domain.SecurityEvent{
			OrganizationID:    orgID,
			SourceID:          sourceUUID,
			SourceEventID:     &e.SourceEventID,
			Source:            "",
			EventType:         e.EventType,
			EventCategory:     types.EventCategory(e.EventCategory),
			Severity:          &e.Severity,
			ActorEmail:        &e.ActorEmail,
			ActorUsername:     &e.ActorUsername,
			IPAddress:         &e.IPAddress,
			GeoCountry:        &e.GeoCountry,
			GeoCity:           &e.GeoCity,
			RawPayload:        e.RawPayload,
			NormalizedPayload: db.JSONMap{"_ingest_type": "api_stream"},
			ParseStatus:       types.ParseStatusPending,
			ParseErrors:       make([]db.JSONMap, 0),
			OccurredAt:        occurredAt,
		}
		events = append(events, event)
	}
	if err := s.eventRepo.BulkCreateEvents(ctx, events); err != nil {
		return nil, err
	}
	// Update source metrics asynchronously
	count := int64(len(events))
	go func() {
		_ = s.dataSourceRepo.IncrementEventsToday(ctx, sourceUUID)
		_ = s.dataSourceRepo.UpdateHealthMetrics(ctx, sourceUUID, count, count, 0, nil, nil)
	}()
	return map[string]interface{}{
		"ingested":        len(events),
		"source_id":       sourceUUID.String(),
		"organization_id": orgID.String(),
	}, nil
}

func (s *LogsService) SearchLogs(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	return s.eventRepo.SearchEvents(ctx, orgID, filters, page, pageSize)
}

func (s *LogsService) GetLogByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*domain.SecurityEvent, error) {
	return s.eventRepo.GetEventByID(ctx, id, orgID)
}

func (s *LogsService) GetThreatsSummary(ctx context.Context, orgID uuid.UUID) (*domain.ThreatsSummary, error) {
	return s.eventRepo.GetThreatsSummary(ctx, orgID)
}
