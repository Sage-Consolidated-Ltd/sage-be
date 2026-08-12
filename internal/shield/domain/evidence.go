package domain

import (
	"time"

	"sage-backend/internal/shared/types"
)

// ContributingEventSummary summarizes an event that contributed to a detection (especially stateful correlation).
type ContributingEventSummary struct {
	EventID       string         `json:"event_id"`
	SourceEventID string         `json:"source_event_id,omitempty"`
	EventType     string         `json:"event_type"`
	OccurredAt    time.Time      `json:"occurred_at"`
	ActorUsername string         `json:"actor_username,omitempty"`
	IPAddress     string         `json:"ip_address,omitempty"`
	Summary       string         `json:"summary,omitempty"`
	RawPayload    map[string]any `json:"raw_payload,omitempty"`
}

// Evidence provides structured, contextual proof explaining why an Incident was created.
type Evidence struct {
	RuleID             string                     `json:"rule_id"`
	RuleName           string                     `json:"rule_name"`
	Severity           types.Severity             `json:"severity"`
	Description        string                     `json:"description"`
	TriggerEventID     string                     `json:"trigger_event_id"`
	TargetUser         string                     `json:"target_user,omitempty"`
	SourceIP           string                     `json:"source_ip,omitempty"`
	HostName           string                     `json:"host_name,omitempty"`
	OccurredAt         time.Time                  `json:"occurred_at"`
	AttemptCount       int                        `json:"attempt_count,omitempty"`
	TimeWindowSeconds  int                        `json:"time_window_seconds,omitempty"`
	ContributingEvents []ContributingEventSummary `json:"contributing_events,omitempty"`
	Context            map[string]interface{}     `json:"context,omitempty"`
}

// AddContext adds key-value metadata to the evidence context.
func (e *Evidence) AddContext(key string, value interface{}) {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
}
