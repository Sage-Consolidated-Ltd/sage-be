package outbound

import (
	"context"
	"sage-backend/internal/organization/domain"

	"github.com/google/uuid"
)

type DashboardRepository interface {
	GetSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
	SaveSnapshot(ctx context.Context, snapshot *domain.OrganizationDashboard) error
	ComputeSnapshot(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
}
