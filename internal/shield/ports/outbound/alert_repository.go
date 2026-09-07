package outbound

import (
	"context"

	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

// AlertRepository defines the outbound contract for persisting and querying threat alerts.
type AlertRepository interface {
	SaveAlert(ctx context.Context, alert *domain.Alert) error
	BulkSaveAlerts(ctx context.Context, alerts []*domain.Alert) error
	ListAlerts(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]*domain.Alert, int, error)
	GetAlertByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Alert, error)
}

