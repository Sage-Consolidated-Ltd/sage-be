package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
)

type AnalysisRepositoryInt interface {
	RecordAnalysis(ctx context.Context, params *models.CreateAnalysisParams) (*models.AnalysisResult, error)
	GetByLogFileID(ctx context.Context, logFileID uuid.UUID) (*models.AnalysisResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error)
	GetThreatsByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]models.Threat, error)
}

type AnalysisRepository struct {
	db *db.DB
}

func NewAnalysisRepository(db *db.DB) AnalysisRepositoryInt {
	return &AnalysisRepository{
		db: db,
	}
}

func(r *AnalysisRepository) RecordAnalysis(ctx context.Context, params *models.CreateAnalysisParams) (*models.AnalysisResult, error) {
	summaryJSON, err := json.Marshal(params.Summary)
	if err != nil {
		return nil, fmt.Errorf("marshal summary: %w", err)
	}

	outcomeJSON, err := json.Marshal(params.Outcome)
	if err != nil {
		return nil, fmt.Errorf("marshal outcome: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert analysis result
	const insertAnalysis = `
		INSERT INTO analysis_results (
			log_file_id, json_input_id, request_type,
			log_type, approach, overall, summary, outcome
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, log_file_id, json_input_id, request_type,
		          log_type, approach, overall, summary, outcome, created_at`

	var result models.AnalysisResult
	err = tx.QueryRowxContext(ctx, insertAnalysis,
		params.LogFileID,
		params.JsonInputID,
		params.RequestType,
		params.LogType,
		params.Approach,
		params.Overall,
		summaryJSON,
		outcomeJSON,
	).StructScan(&result)
	if err != nil {
		return nil, fmt.Errorf("insert analysis: %w", err)
	}

	// 2. Insert threats
	for _, t := range params.Threats {
		evidenceJSON, err := json.Marshal(t.Evidence)
		if err != nil {
			return nil, fmt.Errorf("marshal evidence: %w", err)
		}

		const insertThreat = `
			INSERT INTO threats (
				analysis_id, organization_id, source, title, category,
				severity, mitre, event_count, time_range,
				what_happened, evidence, recommendation
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err = tx.ExecContext(ctx, insertThreat,
			result.ID,
			params.OrganizationID,
			t.Source,
			t.Title,
			t.Category,
			t.Severity,
			t.Mitre,
			t.EventCount,
			t.TimeRange,
			t.WhatHappened,
			evidenceJSON,
			t.Recommendation,
		)
		if err != nil {
			return nil, fmt.Errorf("insert threat [%s]: %w", t.Title, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// attach threats to result before returning
	result.Threats = params.Threats

	return &result, nil
}
func (r *AnalysisRepository) GetByLogFileID(ctx context.Context, logFileID uuid.UUID) (*models.AnalysisResult, error) {
	const q = `SELECT * FROM analysis_results WHERE log_file_id = $1`

	var result models.AnalysisResult
	if err := r.db.QueryRowxContext(ctx, q, logFileID).StructScan(&result); err != nil {
		return nil, err
	}

	threats, err := r.GetThreatsByAnalysisID(ctx, result.ID)
	if err != nil {
		return nil, err
	}

	result.Threats = threats
	return &result, nil
}
func (r *AnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error) {
	const q = `SELECT * FROM analysis_results WHERE id = $1`

	var result models.AnalysisResult
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(&result); err != nil {
		return nil, err
	}

	threats, err := r.GetThreatsByAnalysisID(ctx, result.ID)
	if err != nil {
		return nil, err
	}

	result.Threats = threats
	return &result, nil
}
func (r *AnalysisRepository) GetThreatsByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]models.Threat, error) {
	const q = `SELECT * FROM threats WHERE analysis_id = $1 ORDER BY severity, created_at`

	var threats []models.Threat
	if err := r.db.SelectContext(ctx, &threats, q, analysisID); err != nil {
		return nil, err
	}

	return threats, nil
}