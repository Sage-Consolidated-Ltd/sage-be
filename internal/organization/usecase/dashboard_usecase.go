package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sage-backend/internal/organization/domain"
	"sage-backend/internal/organization/ports/inbound"
	"sage-backend/internal/organization/ports/outbound"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type DashboardBroadcaster func(orgID string, payload interface{})

type DashboardService struct {
	repo        outbound.DashboardRepository
	redisClient *redis.Client
	broadcaster DashboardBroadcaster
}

func NewDashboardService(
	repo outbound.DashboardRepository,
	redisClient *redis.Client,
	broadcaster DashboardBroadcaster,
) inbound.DashboardUseCase {
	return &DashboardService{
		repo:        repo,
		redisClient: redisClient,
		broadcaster: broadcaster,
	}
}

func (s *DashboardService) GetDashboard(
	ctx context.Context,
	orgID uuid.UUID,
	tab domain.DashboardTab,
) (*domain.OrganizationDashboard, error) {
	cacheKey := fmt.Sprintf("org:%s:dashboard:snapshot", orgID.String())

	// 1. Try serving from Redis cache for sub-millisecond response time
	if s.redisClient != nil {
		if cached, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			var d domain.OrganizationDashboard
			if err := json.Unmarshal([]byte(cached), &d); err == nil {
				return filterByTab(&d, tab), nil
			}
		}
	}

	// 2. Fetch materialized snapshot from database (or computes live if first run)
	dashboard, err := s.repo.GetSnapshot(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("fetch dashboard snapshot: %w", err)
	}

	// 3. Cache snapshot in Redis with a 2-minute TTL
	if s.redisClient != nil && dashboard != nil {
		if bytes, err := json.Marshal(dashboard); err == nil {
			_ = s.redisClient.Set(ctx, cacheKey, string(bytes), 2*time.Minute).Err()
		}
	}

	return filterByTab(dashboard, tab), nil
}

func (s *DashboardService) RefreshDashboard(
	ctx context.Context,
	orgID uuid.UUID,
) (*domain.OrganizationDashboard, error) {
	// Recompute fresh snapshot from OLTP tables
	snapshot, err := s.repo.ComputeSnapshot(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("compute fresh snapshot: %w", err)
	}

	// Warm Redis cache immediately
	cacheKey := fmt.Sprintf("org:%s:dashboard:snapshot", orgID.String())
	if s.redisClient != nil && snapshot != nil {
		if bytes, err := json.Marshal(snapshot); err == nil {
			_ = s.redisClient.Set(ctx, cacheKey, string(bytes), 2*time.Minute).Err()
		}
	}

	// Push real-time update to all subscribed WebSockets
	if s.broadcaster != nil {
		s.broadcaster(orgID.String(), snapshot)
	}

	return snapshot, nil
}

// filterByTab returns only the components relevant to the requested dashboard tab, preventing overfetching.
func filterByTab(d *domain.OrganizationDashboard, tab domain.DashboardTab) *domain.OrganizationDashboard {
	if d == nil {
		return nil
	}

	clone := *d
	clone.Tab = tab

	switch tab {
	case domain.TabOverview, "":
		// Overview returns all components
		return &clone

	case domain.TabIdentity:
		return &domain.OrganizationDashboard{
			OrganizationID:  d.OrganizationID,
			Tab:             domain.TabIdentity,
			SecurityScore:   d.SecurityScore,
			IdentityHealth:  d.IdentityHealth,
			ActiveIncidents: d.ActiveIncidents,
			UpdatedAt:       d.UpdatedAt,
		}

	case domain.TabAssets:
		return &domain.OrganizationDashboard{
			OrganizationID:   d.OrganizationID,
			Tab:              domain.TabAssets,
			SecurityScore:    d.SecurityScore,
			EndpointCoverage: d.EndpointCoverage,
			DangerousThreats: d.DangerousThreats,
			GeoThreats:       d.GeoThreats,
			UpdatedAt:        d.UpdatedAt,
		}

	case domain.TabHealth:
		return &domain.OrganizationDashboard{
			OrganizationID:  d.OrganizationID,
			Tab:             domain.TabHealth,
			SecurityScore:   d.SecurityScore,
			Vulnerabilities: d.Vulnerabilities,
			ThreatIntel:     d.ThreatIntel,
			ComplianceRisks: d.ComplianceRisks,
			ThreatTrends:    d.ThreatTrends,
			UpdatedAt:       d.UpdatedAt,
		}

	default:
		return &clone
	}
}
