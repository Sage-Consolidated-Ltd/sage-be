package usecase

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

func TestDataQualityEngine_ValidEvent(t *testing.T) {
	engine := NewDataQualityEngine()
	ctx := context.Background()

	eventID := "4624"
	user := "admin_user"
	ip := "192.168.1.50"
	severity := types.SeverityHigh

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &eventID,
		Source:         "SecurityLog",
		EventType:      "4624",
		EventCategory:  types.EventCategoryAuthentication,
		Severity:       &severity,
		ActorUsername:  &user,
		IPAddress:      &ip,
		ParseStatus:    types.ParseStatusSuccess,
		OccurredAt:     time.Now(),
		NormalizedPayload: db.JSONMap{
			"event_id":     "4624",
			"channel":      "Security",
			"provider":     "Microsoft-Windows-Security-Auditing",
			"computer":     "DC-01.corp.local",
			"user":         user,
			"ip_address":   ip,
			"process_name": "C:\\Windows\\System32\\lsass.exe",
			"user_sid":     "S-1-5-21-3623811015-3361044348-30300820-1013",
		},
	}

	result, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected error evaluating quality: %v", err)
	}

	if result.Score < 80.0 {
		t.Errorf("Expected score >= 80.0 for valid event, got %.2f", result.Score)
	}
	if result.Status != types.QualityGood {
		t.Errorf("Expected QualityGood status, got %s", result.Status)
	}
	if !result.IsAcceptable() {
		t.Errorf("Expected IsAcceptable to be true")
	}
}

func TestDataQualityEngine_MissingFieldsAndMalformedIP(t *testing.T) {
	engine := NewDataQualityEngine()
	ctx := context.Background()

	badIP := "999.888.777.666"

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  nil, // Missing event ID
		Source:         "",   // Missing source
		IPAddress:      &badIP,
		ParseStatus:    types.ParseStatusFailed, // Parse failed
		OccurredAt:     time.Time{},             // Zero timestamp
		NormalizedPayload: db.JSONMap{
			"user_sid": "INVALID_SID_STRING",
		},
	}

	result, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected error evaluating quality: %v", err)
	}

	if result.Score >= 50.0 {
		t.Errorf("Expected severe penalty score (< 50.0), got %.2f", result.Score)
	}
	if result.Status != types.QualityError {
		t.Errorf("Expected QualityError status, got %s", result.Status)
	}
	if len(result.Issues) == 0 {
		t.Errorf("Expected detected quality issues, got 0")
	}
}

func TestDataQualityEngine_FutureTimestamp(t *testing.T) {
	engine := NewDataQualityEngine()
	ctx := context.Background()

	eventID := "4625"
	futureTime := time.Now().Add(24 * time.Hour)

	event := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourceEventID:  &eventID,
		Source:         "SecurityLog",
		OccurredAt:     futureTime,
		ParseStatus:    types.ParseStatusSuccess,
	}

	result, err := engine.EvaluateEvent(ctx, event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundFutureIssue := false
	for _, issue := range result.Issues {
		if issue.Code == "FUTURE_TIMESTAMP" {
			foundFutureIssue = true
			if issue.Severity != domain.QualitySeverityCritical {
				t.Errorf("Expected FUTURE_TIMESTAMP severity to be Critical, got %s", issue.Severity)
			}
		}
	}

	if !foundFutureIssue {
		t.Errorf("Expected FUTURE_TIMESTAMP issue to be detected")
	}
}

func TestDataQualityEngine_DuplicateDetection(t *testing.T) {
	engine := NewDataQualityEngine()
	ctx := context.Background()

	eventID := "4624"
	occurredAt := time.Now()

	event1 := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.UUID{},
		SourceEventID:  &eventID,
		Source:         "Security",
		OccurredAt:     occurredAt,
		ParseStatus:    types.ParseStatusSuccess,
	}

	event2 := &domain.SecurityEvent{
		ID:             uuid.New(),
		OrganizationID: uuid.UUID{},
		SourceEventID:  &eventID,
		Source:         "Security",
		OccurredAt:     occurredAt,
		ParseStatus:    types.ParseStatusSuccess,
	}

	res1, _ := engine.EvaluateEvent(ctx, event1)
	res2, _ := engine.EvaluateEvent(ctx, event2)

	hasDuplicateIssue1 := false
	for _, issue := range res1.Issues {
		if issue.Code == "DUPLICATE_EVENT_DETECTED" {
			hasDuplicateIssue1 = true
		}
	}
	if hasDuplicateIssue1 {
		t.Errorf("First event should not be flagged as duplicate")
	}

	hasDuplicateIssue2 := false
	for _, issue := range res2.Issues {
		if issue.Code == "DUPLICATE_EVENT_DETECTED" {
			hasDuplicateIssue2 = true
		}
	}
	if !hasDuplicateIssue2 {
		t.Errorf("Second identical event should be flagged as DUPLICATE_EVENT_DETECTED")
	}
}
