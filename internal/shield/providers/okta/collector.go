package okta

import (
	"context"

	"sage-backend/internal/shield/models"
)

func (p *OktaProvider) Collect(
	ctx context.Context,
) ([]models.NormalizedEvent, error) {

	signIns, err := p.PollAuditLogs(ctx)
	if err != nil {
		return nil, err
	}

	var events []models.NormalizedEvent

	for _, e := range signIns {

		raw := map[string]interface{}{
			"eventId":   e.EventID,
			"published": e.Published,
			"actor":     e.Actor,
			"client":    e.Client,
			"outcome":   e.Outcome,
			"target":    e.Target,
		}

		var appName string
		if len(e.Target) > 0 {
			appName =
				e.Target[0].
					DisplayName
		}

		events = append(
			events,
			models.NormalizedEvent{
				ID:          e.EventID,
				Provider:    "okta",
				EventType:   "sign_in",
				UserID:      e.Actor.ID,
				UserName:    e.Actor.DisplayName,
				IPAddress:   e.Client.IPAddress,
				Application: appName,
				Timestamp:   e.Published,
				Status:      e.Outcome.Result,
				Raw:         raw,
			},
		)
	}

	return events, nil
}
