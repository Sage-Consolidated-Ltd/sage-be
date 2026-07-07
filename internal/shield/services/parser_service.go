package services

import (
	"context"
	"encoding/json"
	"time"

	// "sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"

	"github.com/google/uuid"
)

type ParserServiceInt interface {
	GetParserSummary(ctx context.Context, orgID uuid.UUID) (total, active int, errorRate float64, lastUpdated *time.Time, err error)
	ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.Parser, int, error)
	GetParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error)
	CreateParser(ctx context.Context, parser *models.Parser) error
	UpdateParser(ctx context.Context, parser *models.Parser, changeNote string, changedBy uuid.UUID) error
	TestParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, sampleLog string, rawPayload map[string]interface{}) (*models.ParserTestResponse, error)
	PreviewParser(ctx context.Context, parserType types.ParserType, logic map[string]interface{}, mappings []map[string]interface{}, sampleLog string, rawPayload map[string]interface{}) (*models.ParserTestResponse, error)
	EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	ValidateParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error)
	ValidateAllParsers(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	ImportParser(ctx context.Context, parser *models.Parser) error
	ExportParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error)
	ListSampleLogs(ctx context.Context, sourceID, parserID *uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]*models.SecurityEvent, int, error)
}

type ParserService struct {
	parserRepo repositories.ParserRepositoryInt
	eventRepo  repositories.SecurityEventRepositoryInt
	sourceRepo repositories.DataSourceRepositoryInt
	jobRepo    repositories.IngestionJobRepositoryInt
}

func NewParserService(
	parserRepo repositories.ParserRepositoryInt,
	eventRepo repositories.SecurityEventRepositoryInt,
	sourceRepo repositories.DataSourceRepositoryInt,
	jobRepo repositories.IngestionJobRepositoryInt,
) ParserServiceInt {
	return &ParserService{
		parserRepo: parserRepo,
		eventRepo:  eventRepo,
		sourceRepo: sourceRepo,
		jobRepo:    jobRepo,
	}
}

func (s *ParserService) GetParserSummary(ctx context.Context, orgID uuid.UUID) (int, int, float64, *time.Time, error) {
	total, active, errorRate, lastUpdated, err := s.parserRepo.GetParserSummary(ctx, orgID)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	return total, active, errorRate, lastUpdated, nil
}

func (s *ParserService) ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.Parser, int, error) {
	return s.parserRepo.ListParsers(ctx, orgID, filters, page, pageSize)
}

func (s *ParserService) GetParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error) {
	return s.parserRepo.GetParserByID(ctx, id, orgID)
}

func (s *ParserService) CreateParser(ctx context.Context, parser *models.Parser) error {
	// Basic validation
	// if parser.ParserType == "" {
	// 	return apperrors.BadException("PARSER_TYPE_REQUIRED")
	// }
	if parser.Status == "" {
		parser.Status = types.ParserStatusActive
	}
	if parser.Tags == nil {
		parser.Tags = []string{}
	}
	if parser.Logic == nil {
		parser.Logic = make(json.RawMessage, 0)
	}
	if parser.Mappings == nil {
		parser.Mappings = json.RawMessage{}
	}
	if parser.SourceID != nil {
		// validate data source exists and belongs to org
		_, err := s.sourceRepo.GetDataSourceByID(ctx, *parser.SourceID, parser.OrganizationID)
		if err != nil {
			return err
		}

		// update parser_id in data source
	}
	// Create initial version
	if err := s.parserRepo.CreateParser(ctx, parser); err != nil {
		return err
	}
	// Create version entry (version 1)
	version := &models.ParserVersion{
		OrganizationID: parser.OrganizationID,
		ParserID:       parser.ID,
		VersionNumber:  1,
		Logic:          parser.Logic,
		Mappings:       parser.Mappings,
	}
	return s.parserRepo.CreateParserVersion(ctx, version)
}

func (s *ParserService) UpdateParser(ctx context.Context, parser *models.Parser, changeNote string, changedBy uuid.UUID) error {
	// Ensure parser exists
	existing, err := s.parserRepo.GetParserByID(ctx, parser.ID, parser.OrganizationID)
	if err != nil {
		return err
	}
	// Determine next version number
	versions, _ := s.parserRepo.GetParserVersions(ctx, parser.ID, parser.OrganizationID)
	nextVersion := len(versions) + 1
	// Create version with existing state before update
	oldVersion := &models.ParserVersion{
		OrganizationID: existing.OrganizationID,
		ParserID:       existing.ID,
		VersionNumber:  nextVersion,
		Logic:          existing.Logic,
		Mappings:       existing.Mappings,
		ChangedBy:      &changedBy,
		ChangeNote:     &changeNote,
	}
	_ = s.parserRepo.CreateParserVersion(ctx, oldVersion)
	// Apply update
	if err := s.parserRepo.UpdateParser(ctx, parser); err != nil {
		return err
	}
	return nil
}

func (s *ParserService) TestParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, sampleLog string, rawPayload map[string]interface{}) (*models.ParserTestResponse, error) {
	parser, err := s.parserRepo.GetParserByID(ctx, parserID, orgID)
	if err != nil {
		return nil, err
	}
	// Simulate parsing based on parser type
	// Placeholder: just echo back success with sample output
	parsed := map[string]interface{}{
		"message": "parsed output (simulated)",
		"parser":  string(parser.ParserType),
		"sample":  sampleLog,
	}
	normalized := map[string]interface{}{
		"normalized": "output",
	}
	errors := make([]map[string]interface{}, 0)
	success := true
	// Store test run
	run := &models.ParserTestRun{
		OrganizationID:   orgID,
		ParserID:         &parserID,
		SampleLog:        &sampleLog,
		RawPayload:       rawPayload,
		ParsedOutput:     parsed,
		NormalizedOutput: normalized,
		Errors:           errors,
		Success:          success,
	}
	_ = s.parserRepo.CreateTestRun(ctx, run)

	return &models.ParserTestResponse{
		Success:          success,
		ParsedOutput:     parsed,
		NormalizedOutput: normalized,
		Errors:           errors,
	}, nil
}

func (s *ParserService) PreviewParser(ctx context.Context, parserType types.ParserType, logic map[string]interface{}, mappings []map[string]interface{}, sampleLog string, rawPayload map[string]interface{}) (*models.ParserTestResponse, error) {
	// Similar to TestParser but without persisting
	parsed := map[string]interface{}{
		"message": "preview parsed output",
		"type":    parserType,
	}
	normalized := map[string]interface{}{
		"mappings_applied": len(mappings),
	}
	return &models.ParserTestResponse{
		Success:          true,
		ParsedOutput:     parsed,
		NormalizedOutput: normalized,
		Errors:           []map[string]interface{}{},
		SchemaPreview:    &map[string]interface{}{"example": "schema"},
	}, nil
}

func (s *ParserService) EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return s.parserRepo.EnableParser(ctx, id, orgID)
}

func (s *ParserService) DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return s.parserRepo.DisableParser(ctx, id, orgID)
}

func (s *ParserService) ValidateParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (map[string]interface{}, error) {
	// Queue validation job
	job := &models.IngestionJob{
		OrganizationID: orgID,
		Status:         models.JobStatusQueued,
		JobType:        models.JobTypeValidation,
		Metadata:       map[string]interface{}{"parser_id": id.String()},
	}
	if err := s.jobRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"job_id": job.ID.String(),
		"status": "queued",
	}, nil
}

func (s *ParserService) ValidateAllParsers(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	// Queue one job per parser or a batch job
	job := &models.IngestionJob{
		OrganizationID: orgID,
		Status:         models.JobStatusQueued,
		JobType:        models.JobTypeValidation,
		Metadata:       map[string]interface{}{"scope": "all_parsers"},
	}
	if err := s.jobRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"job_id": job.ID.String(),
		"status": "queued",
	}, nil
}

func (s *ParserService) ImportParser(ctx context.Context, parser *models.Parser) error {
	// Parse JSON definition from possibly external file
	if parser.Status == "" {
		parser.Status = types.ParserStatusActive
	}
	return s.parserRepo.ImportParser(ctx, parser)
}

func (s *ParserService) ExportParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error) {
	return s.parserRepo.GetParserByID(ctx, id, orgID)
}

func (s *ParserService) ListSampleLogs(ctx context.Context, sourceID, parserID *uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]*models.SecurityEvent, int, error) {
	// If parserID provided, get events parsed by that parser
	if parserID != nil {
		events, err := s.eventRepo.GetEventsByParser(ctx, *parserID, orgID, pageSize*2) // fetch more
		if err != nil {
			return nil, 0, err
		}
		// Apply simple pagination manually
		total := len(events)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		return events[start:end], total, nil
	}
	// If sourceID provided, get events from that source
	if sourceID != nil {
		return s.eventRepo.GetEventsBySource(ctx, *sourceID, orgID, nil, page, pageSize)
	}
	// Otherwise general search
	return s.eventRepo.SearchEvents(ctx, orgID, nil, page, pageSize)
}
