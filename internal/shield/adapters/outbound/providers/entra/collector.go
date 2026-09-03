package entra

import (
	"context"
	"sage-backend/internal/shield/domain"
	"time"
)

func (p *EntraProvider) Collect(
	ctx context.Context,
	limit int,
) ([]domain.NormalizedEvent, error) {
	signIns, err := p.PollAuditLogs(
		ctx,
		limit,
	)
	if err != nil {
		return nil, err
	}

	var events []domain.NormalizedEvent

	for _, e := range signIns {
		raw := map[string]interface{}{
			"id":              e.ID,
			"userDisplayName": e.UserDisplayName,
			"userId":          e.UserID,
			"appDisplayName":  e.AppDisplayName,
			"ipAddress":       e.IPAddress,
			"createdDateTime": e.CreatedDateTime,
			"status":          e.Status,
		}

		timestamp, _ := time.Parse(
			time.RFC3339,
			e.CreatedDateTime,
		)

		events = append(
			events,
			domain.NormalizedEvent{
				ID:        e.ID,
				Provider:  "entra",
				EventType: "sign_in",
				UserID:    e.UserID,
				UserName:  e.UserDisplayName,
				IPAddress: e.IPAddress,
				Timestamp: timestamp,
				Raw:       raw,
			},
		)
	}

	return events, nil
}
