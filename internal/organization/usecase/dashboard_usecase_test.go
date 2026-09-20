package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/organization/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDashboardRepo struct {
	getSnapshotFn     func(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
	saveSnapshotFn    func(ctx context.Context, snapshot *domain.OrganizationDashboard) error
	computeSnapshotFn func(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
}

func (m *mockDashboardRepo) GetSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
	if m.getSnapshotFn != nil {
		return m.getSnapshotFn(ctx, orgID)
	}
	return nil, nil
}

func (m *mockDashboardRepo) SaveSnapshot(ctx context.Context, snapshot *domain.OrganizationDashboard) error {
	if m.saveSnapshotFn != nil {
		return m.saveSnapshotFn(ctx, snapshot)
	}
	return nil
}

func (m *mockDashboardRepo) ComputeSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
	if m.computeSnapshotFn != nil {
		return m.computeSnapshotFn(ctx, orgID)
	}
	return nil, nil
}

func sampleDashboard(orgID uuid.UUID) *domain.OrganizationDashboard {
	return &domain.OrganizationDashboard{
		OrganizationID: orgID,
		Tab:            domain.TabOverview,
		SecurityScore: &domain.SecurityScore{
			OverallScore: 94,
			WeeklyDelta:  4,
		},
		Vulnerabilities: &domain.VulnerabilitiesSummary{
			Critical: 8,
			High:     12,
			Medium:   27,
			Low:      63,
			Total:    110,
		},
		IdentityHealth: &domain.IdentityHealthSummary{
			CoveragePercentage: 77,
			ElevatedPrivileges: 23,
		},
		EndpointCoverage: &domain.AssetProtectionCoverage{
			CoveragePercentage: 95,
			ProtectedCount:     1110,
		},
		ThreatIntel: &domain.ThreatIntelFeedsSummary{
			ActiveFeeds:   13,
			InactiveFeeds: 3,
		},
		ActiveIncidents: []domain.ActiveIncident{
			{ID: "inc-1", IncidentName: "Suspicious Login", Severity: "High", Status: "Active"},
		},
		DangerousThreats: &domain.AssetRiskDistribution{
			OverallRiskPercentage: 25,
			Breakdown: []domain.AssetRiskItem{
				{AssetName: "prod-db-cluster-01", Percentage: 25},
			},
		},
		ComplianceRisks: &domain.ComplianceRiskIndicators{
			EncryptionVulnerabilities: 3,
			ExcessiveUserPermissions:  12,
		},
		ThreatTrends: &domain.ThreatTrendsSummary{
			CurrentMonth: "Sep 2026",
			Days: []domain.ThreatDayTrend{
				{Day: 1, Critical: 2, High: 5, Medium: 10, Low: 20, Total: 37},
			},
		},
		GeoThreats: &domain.GeoThreatsSummary{
			TotalThreats:     2439,
			HighThreatRegion: "North America",
			Origins: []domain.GeoThreatOrigin{
				{Country: "United States", Lat: 37.0902, Lng: -95.7129, Count: 890, Percentage: 36.5},
			},
		},
		UpdatedAt: time.Now().UTC(),
	}
}

func TestDashboardService_GetDashboard_FallbackToDB(t *testing.T) {
	orgID := uuid.New()
	expected := sampleDashboard(orgID)

	repo := &mockDashboardRepo{
		getSnapshotFn: func(ctx context.Context, id uuid.UUID) (*domain.OrganizationDashboard, error) {
			assert.Equal(t, orgID, id)
			return expected, nil
		},
	}

	service := NewDashboardService(repo, nil, nil)
	result, err := service.GetDashboard(context.Background(), orgID, domain.TabOverview)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, orgID, result.OrganizationID)
	assert.Equal(t, 94, result.SecurityScore.OverallScore)
	assert.Equal(t, 110, result.Vulnerabilities.Total)
}

func TestDashboardService_RefreshDashboard_Broadcast(t *testing.T) {
	orgID := uuid.New()
	expected := sampleDashboard(orgID)

	var broadcastOrgID string
	var broadcastPayload interface{}

	broadcaster := func(id string, payload interface{}) {
		broadcastOrgID = id
		broadcastPayload = payload
	}

	repo := &mockDashboardRepo{
		computeSnapshotFn: func(ctx context.Context, id uuid.UUID) (*domain.OrganizationDashboard, error) {
			assert.Equal(t, orgID, id)
			return expected, nil
		},
	}

	service := NewDashboardService(repo, nil, broadcaster)
	result, err := service.RefreshDashboard(context.Background(), orgID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, orgID.String(), broadcastOrgID)
	assert.Equal(t, expected, broadcastPayload)
}

func TestDashboardService_FilterByTab(t *testing.T) {
	orgID := uuid.New()
	full := sampleDashboard(orgID)

	t.Run("Overview returns all widgets", func(t *testing.T) {
		filtered := filterByTab(full, domain.TabOverview)
		require.NotNil(t, filtered)
		assert.Equal(t, domain.TabOverview, filtered.Tab)
		assert.NotNil(t, filtered.SecurityScore)
		assert.NotNil(t, filtered.Vulnerabilities)
		assert.NotNil(t, filtered.IdentityHealth)
		assert.NotNil(t, filtered.EndpointCoverage)
		assert.NotNil(t, filtered.ThreatIntel)
		assert.NotEmpty(t, filtered.ActiveIncidents)
		assert.NotNil(t, filtered.DangerousThreats)
		assert.NotNil(t, filtered.ComplianceRisks)
		assert.NotNil(t, filtered.ThreatTrends)
		assert.NotNil(t, filtered.GeoThreats)
	})

	t.Run("Identity tab filters identity-specific data", func(t *testing.T) {
		filtered := filterByTab(full, domain.TabIdentity)
		require.NotNil(t, filtered)
		assert.Equal(t, domain.TabIdentity, filtered.Tab)
		assert.NotNil(t, filtered.SecurityScore)
		assert.NotNil(t, filtered.IdentityHealth)
		assert.NotEmpty(t, filtered.ActiveIncidents)
		assert.Nil(t, filtered.Vulnerabilities)
		assert.Nil(t, filtered.EndpointCoverage)
		assert.Nil(t, filtered.GeoThreats)
	})

	t.Run("Assets tab filters asset and geo-threat data", func(t *testing.T) {
		filtered := filterByTab(full, domain.TabAssets)
		require.NotNil(t, filtered)
		assert.Equal(t, domain.TabAssets, filtered.Tab)
		assert.NotNil(t, filtered.SecurityScore)
		assert.NotNil(t, filtered.EndpointCoverage)
		assert.NotNil(t, filtered.DangerousThreats)
		assert.NotNil(t, filtered.GeoThreats)
		assert.Nil(t, filtered.IdentityHealth)
		assert.Nil(t, filtered.Vulnerabilities)
	})

	t.Run("Health tab filters health and vulnerability trends", func(t *testing.T) {
		filtered := filterByTab(full, domain.TabHealth)
		require.NotNil(t, filtered)
		assert.Equal(t, domain.TabHealth, filtered.Tab)
		assert.NotNil(t, filtered.SecurityScore)
		assert.NotNil(t, filtered.Vulnerabilities)
		assert.NotNil(t, filtered.ThreatIntel)
		assert.NotNil(t, filtered.ComplianceRisks)
		assert.NotNil(t, filtered.ThreatTrends)
		assert.Nil(t, filtered.EndpointCoverage)
		assert.Nil(t, filtered.IdentityHealth)
	})
}
