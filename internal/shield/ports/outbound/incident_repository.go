package outbound

import (
	"context"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

// IncidentRepository defines the outbound contract for persisting and retrieving incidents.
type IncidentRepository interface {
	SaveIncident(ctx context.Context, incident *domain.Incident) error
	GetIncidentByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Incident, error)
	ListIncidents(ctx context.Context, orgID uuid.UUID, status *domain.IncidentStatus, limit, offset int) ([]*domain.Incident, int, error)
	UpdateIncidentStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.IncidentStatus) error
}
