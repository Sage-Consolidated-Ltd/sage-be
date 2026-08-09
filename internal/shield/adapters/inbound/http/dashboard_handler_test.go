package http

import (
	"context"
	"net/http/httptest"
	"sage-backend/internal/shield/domain"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDashboardService struct {
	mock.Mock
}

func (m *mockDashboardService) GetSecurityPostureScore(ctx context.Context, orgID uuid.UUID) (*domain.SecurityPostureScore, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityPostureScore), args.Error(1)
}

func (m *mockDashboardService) GetIdentityHealthSummary(ctx context.Context, orgID uuid.UUID) (*domain.IdentityHealthSummary, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IdentityHealthSummary), args.Error(1)
}

func (m *mockDashboardService) GetAssetProtectionCoverage(ctx context.Context, orgID uuid.UUID) (*domain.AssetProtectionCoverage, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AssetProtectionCoverage), args.Error(1)
}

func (m *mockDashboardService) GetThreatIntelFeedsSummary(ctx context.Context, orgID uuid.UUID) (*domain.ThreatIntelFeedsSummary, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ThreatIntelFeedsSummary), args.Error(1)
}

func (m *mockDashboardService) GetActiveIncidents(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.ActiveIncident, error) {
	args := m.Called(ctx, orgID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ActiveIncident), args.Error(1)
}

func (m *mockDashboardService) GetAssetRiskDistribution(ctx context.Context, orgID uuid.UUID) (*domain.AssetRiskDistribution, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AssetRiskDistribution), args.Error(1)
}

func (m *mockDashboardService) GetComplianceRiskIndicators(ctx context.Context, orgID uuid.UUID) (*domain.ComplianceRiskIndicators, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplianceRiskIndicators), args.Error(1)
}

func (m *mockDashboardService) GetThreatTrends(ctx context.Context, orgID uuid.UUID) (*domain.ThreatTrendsSummary, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ThreatTrendsSummary), args.Error(1)
}

func (m *mockDashboardService) GetGeoThreats(ctx context.Context, orgID uuid.UUID) (*domain.GeoThreatsSummary, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GeoThreatsSummary), args.Error(1)
}

func TestDashboardHandler_Endpoints(t *testing.T) {
	app := fiber.New()
	mockService := new(mockDashboardService)
	handler := NewDashboardHandler(mockService)

	orgID := uuid.New()

	mockService.On("GetSecurityPostureScore", mock.Anything, orgID).Return(&domain.SecurityPostureScore{OverallScore: 94}, nil)
	mockService.On("GetIdentityHealthSummary", mock.Anything, orgID).Return(&domain.IdentityHealthSummary{CoveragePercentage: 77}, nil)
	mockService.On("GetAssetProtectionCoverage", mock.Anything, orgID).Return(&domain.AssetProtectionCoverage{CoveragePercentage: 95}, nil)

	app.Get("/security-posture/score", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetSecurityPostureScore(c)
	})
	app.Get("/identity-health/summary", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetIdentityHealthSummary(c)
	})
	app.Get("/assets/protection-coverage", func(c *fiber.Ctx) error {
		c.Locals("orgID", orgID)
		return handler.GetAssetProtectionCoverage(c)
	})

	tests := []struct {
		url          string
		expectedCode int
	}{
		{"/security-posture/score", 200},
		{"/identity-health/summary", 200},
		{"/assets/protection-coverage", 200},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.url, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedCode, resp.StatusCode)
	}

	mockService.AssertExpectations(t)
}
