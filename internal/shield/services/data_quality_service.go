package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"

	"github.com/google/uuid"
)

type DataQualityServiceInt interface {
	GetSummary(ctx context.Context, orgID uuid.UUID) (*models.DataQualityScan, error)
	RunScan(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	GetBreakdown(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID, page, pageSize int) ([]*models.DataQualitySourceMetric, error)
	GetAIAnalysis(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID) ([]map[string]interface{}, error)
	ApplySuggestedFix(ctx context.Context, orgID uuid.UUID, suggestionID uuid.UUID) error
	GetSuggestedFixDiff(ctx context.Context, suggestionID, parserID uuid.UUID) (map[string]interface{}, error)
	DownloadDataQualityReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error)
}

type DataQualityService struct {
	scanRepo   repositories.DataQualityRepositoryInt
	parserRepo repositories.ParserRepositoryInt
	sourceRepo repositories.DataSourceRepositoryInt
	jobRepo    repositories.IngestionJobRepositoryInt
}

func NewDataQualityService(
	scanRepo repositories.DataQualityRepositoryInt,
	parserRepo repositories.ParserRepositoryInt,
	sourceRepo repositories.DataSourceRepositoryInt,
	jobRepo repositories.IngestionJobRepositoryInt,
) DataQualityServiceInt {
	return &DataQualityService{
		scanRepo:   scanRepo,
		parserRepo: parserRepo,
		sourceRepo: sourceRepo,
		jobRepo:    jobRepo,
	}
}

func (s *DataQualityService) GetSummary(ctx context.Context, orgID uuid.UUID) (*models.DataQualityScan, error) {
	// Return the latest completed scan
	scans, err := s.scanRepo.ListScans(ctx, orgID, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(scans) == 0 {
		// No scan yet, return zero-value summary
		return &models.DataQualityScan{
			QualityScore:              nil,
			ParsingErrors:             nil,
			MissingFieldsPercentage:   nil,
			DuplicateEventsPercentage: nil,
			UnmappedLogsCount:         nil,
		}, nil
	}
	return scans[0], nil
}

func (s *DataQualityService) RunScan(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	scan := &models.DataQualityScan{
		OrganizationID: orgID,
		Status:         "running",
		StartedAt:      time.Now(),
	}
	err := s.scanRepo.CreateScan(ctx, scan)
	if err != nil {
		return nil, err
	}
	// Create job record
	job := &models.IngestionJob{
		OrganizationID: orgID,
		Status:         models.JobStatusQueued,
		JobType:        models.JobTypeQualityScan,
		Metadata:       map[string]interface{}{"scan_id": scan.ID.String()},
	}
	if err := s.jobRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"scan_id":    scan.ID.String(),
		"job_id":     job.ID.String(),
		"status":     "running",
		"started_at": scan.StartedAt,
	}, nil
}

func (s *DataQualityService) GetBreakdown(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID, page, pageSize int) ([]*models.DataQualitySourceMetric, error) {
	if scanID == nil {
		// Use latest scan
		scan, err := s.GetSummary(ctx, orgID)
		if err != nil {
			return nil, err
		}
		if scan == nil || scan.ID == uuid.Nil {
			return []*models.DataQualitySourceMetric{}, nil
		}
		scanID = &scan.ID
	}
	metrics, err := s.scanRepo.GetSourceMetricsByScan(ctx, *scanID, orgID)
	if err != nil {
		return nil, err
	}
	// Optionally enrich with source name
	// Not needed; handler can join.
	return metrics, nil
}

func (s *DataQualityService) GetAIAnalysis(ctx context.Context, orgID uuid.UUID, scanID *uuid.UUID) ([]map[string]interface{}, error) {
	// Get suggestions for the organization or specific scan?
	// Suggestions are stored in data_quality_suggestions.
	// Let's fetch suggestions that are pending or recent.
	suggestions, err := s.scanRepo.GetSuggestions(ctx, orgID, nil, nil, models.SuggestionStatusPending)
	if err != nil {
		return nil, err
	}
	insights := make([]map[string]interface{}, 0, len(suggestions))
	for _, sug := range suggestions {
		insights = append(insights, map[string]interface{}{
			"id":             sug.ID.String(),
			"title":          sug.Summary,
			"message":        sug.Recommendation,
			"recommendation": sug.SuggestedFix,
			"confidence":     sug.Confidence,
			"created_at":     sug.CreatedAt,
		})
	}
	return insights, nil
}

func (s *DataQualityService) ApplySuggestedFix(ctx context.Context, orgID uuid.UUID, suggestionID uuid.UUID) error {
	// Fetch suggestion
	suggestions, err := s.scanRepo.GetSuggestions(ctx, orgID, nil, nil, models.SuggestionStatusPending)
	if err != nil {
		return err
	}
	var target *models.DataQualitySuggestion
	for _, sug := range suggestions {
		if sug.ID == suggestionID {
			target = sug
			break
		}
	}
	if target == nil {
		return apperrors.NotFoundError("SUGGESTION NOT FOUND")
	}
	// Apply fix: if parser_id is set, update parser logic/mappings
	if target.ParserID != nil {
		parser, err := s.parserRepo.GetParserByID(ctx, *target.ParserID, orgID)
		if err != nil {
			return err
		}
		// Update parser fields based on suggested_fix
		// expected suggested_fix contains 'logic' and/or 'mappings'
		if logic, ok := target.SuggestedFix["logic"].(map[string]interface{}); ok {
			parser.Logic = logic
		}
		if mappings, ok := target.SuggestedFix["mappings"].([]map[string]interface{}); ok {
			parser.Mappings = mappings
		}
		// Create version before update
		version := &models.ParserVersion{
			OrganizationID: orgID,
			ParserID:       parser.ID,
			VersionNumber:  1, // increment later
			Logic:          parser.Logic,
			Mappings:       parser.Mappings,
			ChangedBy:      &orgID, // user id? use org for now
		}
		// Get existing version count to set version number
		versions, _ := s.parserRepo.GetParserVersions(ctx, parser.ID, orgID)
		version.VersionNumber = len(versions) + 1
		_ = s.parserRepo.CreateParserVersion(ctx, version)

		// Update parser
		if err := s.parserRepo.UpdateParser(ctx, parser); err != nil {
			return err
		}
	}
	// Mark suggestion as applied
	return s.scanRepo.UpdateSuggestionStatus(ctx, suggestionID, orgID, models.SuggestionStatusApplied)
}

func (s *DataQualityService) GetSuggestedFixDiff(ctx context.Context, suggestionID, parserID uuid.UUID) (map[string]interface{}, error) {
	// Retrieve suggestion
	suggestions, err := s.scanRepo.GetSuggestions(ctx, uuid.Nil, &parserID, nil, models.SuggestionStatusPending)
	if err != nil {
		return nil, err
	}
	var target *models.DataQualitySuggestion
	for _, sug := range suggestions {
		if sug.ID == suggestionID {
			target = sug
			break
		}
	}
	if target == nil {
		return nil, apperrors.NotFoundError("SUGGESTION NOT FOUND")
	}
	// Get current parser
	orgID := target.OrganizationID // can get from suggestion
	parser, err := s.parserRepo.GetParserByID(ctx, parserID, orgID)
	if err != nil {
		return nil, err
	}
	// Compute diff: simple before/after maps
	before := map[string]interface{}{
		"logic":    parser.Logic,
		"mappings": parser.Mappings,
	}
	after := target.SuggestedFix
	// Identify changed fields
	changed := []string{}
	if _, ok := target.SuggestedFix["logic"]; ok {
		changed = append(changed, "logic")
	}
	if _, ok := target.SuggestedFix["mappings"]; ok {
		changed = append(changed, "mappings")
	}
	return map[string]interface{}{
		"before":         before,
		"after":          after,
		"changed_fields": changed,
	}, nil
}

func (s *DataQualityService) DownloadDataQualityReport(ctx context.Context, orgID uuid.UUID, format string, startTime, endTime *time.Time) ([]byte, string, error) {
	// Get summary data
	summary, err := s.GetSummary(ctx, orgID)
	if err != nil {
		return nil, "", err
	}

	// Get breakdown data
	metrics, err := s.GetBreakdown(ctx, orgID, nil, 1, 1000)
	if err != nil {
		return nil, "", err
	}

	// Get AI analysis
	insights, err := s.GetAIAnalysis(ctx, orgID, nil)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("data-quality-report-%s.%s", time.Now().Format("20060102-150405"), format)

	// Generate report based on format
	switch format {
	case "json":
		report := map[string]interface{}{
			"summary":      summary,
			"breakdown":    metrics,
			"insights":     insights,
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
		csv.WriteString("Source ID,Parsing Errors,Missing Fields %,Unmapped Events,Duplicate %,Status\n")
		for _, m := range metrics {
			csv.WriteString(fmt.Sprintf("%s,%d,%.2f,%d,%.2f,%s\n",
				m.SourceID,
				m.ParsingErrors,
				m.MissingFieldsPercentage,
				m.UnmappedEvents,
				m.DuplicatePercentage,
				m.Status))
		}
		return []byte(csv.String()), filename, nil
	case "pdf":
		// Placeholder for PDF generation - would require a PDF library
		// For now, return JSON as fallback
		report := map[string]interface{}{
			"summary":      summary,
			"breakdown":    metrics,
			"insights":     insights,
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
