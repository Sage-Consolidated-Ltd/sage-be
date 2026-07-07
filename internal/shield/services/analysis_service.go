package services

import (
	"context"
	"fmt"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/repositories"

	"github.com/google/uuid"
)

type AnalysisServiceInt interface {
	RecordAnalysis(ctx context.Context, params models.CreateAnalysisParams) (*models.AnalysisResult, error)
	GetByLogFileID(ctx context.Context, logFileID uuid.UUID) (*models.AnalysisResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error)
	GetThreatsByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]models.Threat, error)
}

type AnalysisService struct {
	analysisRepo repositories.AnalysisRepositoryInt
}

func NewAnalysisService(analysisRepo repositories.AnalysisRepositoryInt) AnalysisServiceInt {
	return &AnalysisService{
		analysisRepo: analysisRepo,
	}
}

func (s *AnalysisService) RecordAnalysis(ctx context.Context, params models.CreateAnalysisParams) (*models.AnalysisResult, error) {
	result, err := s.analysisRepo.RecordAnalysis(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("record analysis: %w", err)
	}
	return result, nil
}

func (s *AnalysisService) GetByLogFileID(ctx context.Context, logFileID uuid.UUID) (*models.AnalysisResult, error) {
	result, err := s.analysisRepo.GetByLogFileID(ctx, logFileID)
	if err != nil {
		return nil, fmt.Errorf("get analysis by log file: %w", err)
	}
	return result, nil
}

func (s *AnalysisService) GetByID(ctx context.Context, id uuid.UUID) (*models.AnalysisResult, error) {
	result, err := s.analysisRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get analysis by id: %w", err)
	}
	return result, nil
}

func (s *AnalysisService) GetThreatsByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]models.Threat, error) {
	threats, err := s.analysisRepo.GetThreatsByAnalysisID(ctx, analysisID)
	if err != nil {
		return nil, fmt.Errorf("get threats: %w", err)
	}
	return threats, nil
}
