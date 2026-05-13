package entra

import (
	"context"
	"sage-backend/internal/shield/models"
	"time"
)

func (p *EntraProvider) Collect(
	ctx context.Context,
) ([]models.NormalizedEvent, error) {

	signIns, err := p.PollAuditLogs(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	var events []models.NormalizedEvent

	for _, e := range signIns {

		timestamp, _ := time.Parse(
			time.RFC3339,
			e.CreatedDateTime,
		)

		events = append(
			events,
			models.NormalizedEvent{
				ID:        e.ID,
				Provider:  "entra",
				EventType: "sign_in",
				UserID:    e.UserID,
				UserName:  e.UserDisplayName,
				IPAddress: e.IPAddress,
				Timestamp: timestamp,
			},
		)
	}

	return events, nil
}