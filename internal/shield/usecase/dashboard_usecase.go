package usecase

import (
	"context"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type DashboardService struct {
	repo outbound.DashboardRepository
}

func NewDashboardService(repo outbound.DashboardRepository) inbound.DashboardUseCase {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSecurityPostureScore(ctx context.Context, orgID uuid.UUID) (*domain.SecurityPostureScore, error) {
	return s.repo.GetSecurityPostureScore(ctx, orgID)
}

func (s *DashboardService) GetIdentityHealthSummary(ctx context.Context, orgID uuid.UUID) (*domain.IdentityHealthSummary, error) {
	return s.repo.GetIdentityHealthSummary(ctx, orgID)
}

func (s *DashboardService) GetAssetProtectionCoverage(ctx context.Context, orgID uuid.UUID) (*domain.AssetProtectionCoverage, error) {
	return s.repo.GetAssetProtectionCoverage(ctx, orgID)
}

func (s *DashboardService) GetThreatIntelFeedsSummary(ctx context.Context, orgID uuid.UUID) (*domain.ThreatIntelFeedsSummary, error) {
	return s.repo.GetThreatIntelFeedsSummary(ctx, orgID)
}

func (s *DashboardService) GetActiveIncidents(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.ActiveIncident, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetActiveIncidents(ctx, orgID, limit)
}

func (s *DashboardService) GetAssetRiskDistribution(ctx context.Context, orgID uuid.UUID) (*domain.AssetRiskDistribution, error) {
	return s.repo.GetAssetRiskDistribution(ctx, orgID)
}

func (s *DashboardService) GetComplianceRiskIndicators(ctx context.Context, orgID uuid.UUID) (*domain.ComplianceRiskIndicators, error) {
	return s.repo.GetComplianceRiskIndicators(ctx, orgID)
}

func (s *DashboardService) GetThreatTrends(ctx context.Context, orgID uuid.UUID, currentMonth, previousMonth string) (*domain.ThreatTrendsSummary, error) {
	return s.repo.GetThreatTrends(ctx, orgID, currentMonth, previousMonth)
}

func (s *DashboardService) GetGeoThreats(ctx context.Context, orgID uuid.UUID) (*domain.GeoThreatsSummary, error) {
	return s.repo.GetGeoThreats(ctx, orgID)
}
