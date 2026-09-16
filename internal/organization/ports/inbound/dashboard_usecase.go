package inbound

import (
	"context"
	"sage-backend/internal/organization/domain"

	"github.com/google/uuid"
)

type DashboardUseCase interface {
	GetDashboard(ctx context.Context, orgID uuid.UUID, tab domain.DashboardTab) (*domain.OrganizationDashboard, error)
	RefreshDashboard(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
}
