package models

import (
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type AnalysisResultDTO struct {
	ID          uuid.UUID                              `db:"id"`
	LogFileID   *uuid.UUID                             `db:"log_file_id"`
	JsonInputID *uuid.UUID                             `db:"json_input_id"`
	RequestType string                                 `db:"request_type"`
	LogType     string                                 `db:"log_type"`
	Approach    string                                 `db:"approach"`
	Overall     string                                 `db:"overall"`
	Summary     db.GenericJSON[domain.AnalysisSummary] `db:"summary"`
	Outcome     db.GenericJSON[domain.AnalysisOutcome] `db:"outcome"`
	CreatedAt   time.Time                              `db:"created_at"`
}

func (dto *AnalysisResultDTO) ToDomain() *domain.AnalysisResult {
	if dto == nil {
		return nil
	}

	return &domain.AnalysisResult{
		ID:          dto.ID,
		LogFileID:   dto.LogFileID,
		JsonInputID: dto.JsonInputID,
		RequestType: domain.AnalysisRequestType(dto.RequestType),
		LogType:     dto.LogType,
		Approach:    dto.Approach,
		Overall:     dto.Overall,
		Summary:     dto.Summary,
		Outcome:     dto.Outcome,
		CreatedAt:   dto.CreatedAt,
	}
}

type ThreatDTO struct {
	ID             uuid.UUID                `db:"id"`
	AnalysisID     uuid.UUID                `db:"analysis_id"`
	OrganizationID *uuid.UUID               `db:"organization_id"`
	Source         string                   `db:"source"`
	Title          string                   `db:"title"`
	Category       string                   `db:"category"`
	Severity       string                   `db:"severity"`
	Mitre          string                   `db:"mitre"`
	EventCount     int                      `db:"event_count"`
	TimeRange      string                   `db:"time_range"`
	WhatHappened   string                   `db:"what_happened"`
	Evidence       db.GenericJSON[[]string] `db:"evidence"`
	Recommendation string                   `db:"recommendation"`
	CreatedAt      time.Time                `db:"created_at"`
}

func (dto *ThreatDTO) ToDomain() domain.Threat {
	return domain.Threat{
		ID:             dto.ID,
		AnalysisID:     dto.AnalysisID,
		Source:         dto.Source,
		Title:          dto.Title,
		Category:       dto.Category,
		Severity:       dto.Severity,
		Mitre:          dto.Mitre,
		EventCount:     dto.EventCount,
		TimeRange:      dto.TimeRange,
		WhatHappened:   dto.WhatHappened,
		Evidence:       dto.Evidence.Data,
		Recommendation: dto.Recommendation,
	}
}
