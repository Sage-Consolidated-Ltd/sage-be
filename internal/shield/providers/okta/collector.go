package okta

import (
	"context"
	"time"

	"sage-backend/internal/shield/models"
)

func (p *OktaProvider) Collect(
	ctx context.Context,
	limit int,
) ([]models.NormalizedEvent, error) {

	signIns, err := p.PollAuditLogs(ctx, limit)
	if err != nil {
		return nil, err
	}

	var events []models.NormalizedEvent

	for _, e := range signIns {

		actor := getMap(e, "actor")
		// client := getMap(e, "client")
		outcome := getMap(e, "outcome")
		request := getMap(e, "request")

		// actor fields
		userID := getString(actor, "id")
		userName := getString(actor, "displayName")

		// event type
		eventType := getString(e, "eventType")

		// outcome
		status := ""
		if outcome != nil {
			status = getString(outcome, "result")
		}

		// IP extraction (nested)
		var ip string
		if request != nil {
			ipChain := getArray(request, "ipChain")
			if len(ipChain) > 0 {
				if first, ok := ipChain[0].(map[string]interface{}); ok {
					ip = getString(first, "ip")
				}
			}
		}

		// application
		var appName string
		if targets := getArray(e, "target"); len(targets) > 0 {
			if t0, ok := targets[0].(map[string]interface{}); ok {
				appName = getString(t0, "displayName")
			}
		}

		events = append(events, models.NormalizedEvent{
			ID:          getString(e, "uuid"),
			Provider:    "okta",
			EventType:   eventType,
			UserID:      userID,
			UserName:    userName,
			IPAddress:   ip,
			Application: appName,
			Status:      status,
			Timestamp:   parseTime(getString(e, "published")),
			Raw:         e,
		})
	}

	return events, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

func getArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	return nil
}

func parseTime(t string) time.Time {
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
