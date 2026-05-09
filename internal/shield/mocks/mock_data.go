package mocks

import (
	"encoding/json"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"
	"time"

	"github.com/google/uuid"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// GenerateMockSecurityEvent creates a mock security event for testing
func GenerateMockSecurityEvent(orgID uuid.UUID) *models.SecurityEvent {
	id := uuid.New()
	sourceID := uuid.New()
	eventID := "EVT-" + id.String()[:8]
	actorEmail := "test@example.com"
	actorUsername := "testuser"
	ipAddress := "192.168.1.1"

	return &models.SecurityEvent{
		ID:                id,
		OrganizationID:    orgID,
		SourceID:          sourceID,
		SourceEventID:     &eventID,
		Source:            "microsoft_365",
		EventType:         "user_login",
		EventCategory:     types.EventCategoryAuthentication,
		Severity:          types.SeverityMedium,
		ActorEmail:        &actorEmail,
		ActorUsername:     &actorUsername,
		IPAddress:         &ipAddress,
		RawPayload:        map[string]interface{}{"event_id": id.String(), "action": "login"},
		NormalizedPayload: map[string]interface{}{"user": "testuser", "action": "login_success"},
		ParseStatus:       types.ParseStatusSuccess,
		OccurredAt:        time.Now().Add(-1 * time.Hour),
		IngestedAt:        time.Now(),
		CreatedAt:         time.Now(),
	}
}

// GenerateMockSecurityEvents creates multiple mock security events
func GenerateMockSecurityEvents(orgID uuid.UUID, count int) []models.SecurityEvent {
	events := make([]models.SecurityEvent, count)
	eventTypes := []string{"user_login", "file_access", "email_sent", "permission_change"}
	categories := []types.EventCategory{
		types.EventCategoryAuthentication,
		types.EventCategoryAccess,
		types.EventCategoryNetwork,
		types.EventCategorySystem,
	}
	severities := []types.Severity{types.SeverityLow, types.SeverityMedium, types.SeverityHigh, types.SeverityCritical}

	for i := 0; i < count; i++ {
		event := GenerateMockSecurityEvent(orgID)
		event.EventType = eventTypes[i%len(eventTypes)]
		event.EventCategory = categories[i%len(categories)]
		event.Severity = severities[i%len(severities)]
		events[i] = *event
	}
	return events
}

// GenerateMockDataSource creates a mock data source for testing
func GenerateMockDataSource(orgID uuid.UUID) *models.DataSource {
	id := uuid.New()
	lastSyncAt := time.Now().Add(-5 * time.Minute)
	lastEventAt := time.Now().Add(-10 * time.Minute)

	return &models.DataSource{
		ID:               id,
		OrganizationID:   orgID,
		Name:             "Microsoft 365",
		Type:             "microsoft_365",
		Provider:         strPtr("microsoft"),
		Status:           models.DataSourceStatusActive,
		EventsToday:      1250,
		TotalEvents:      50000,
		LastEventAt:      &lastEventAt,
		LastSyncAt:       &lastSyncAt,
		ErrorCount:       3,
		DelayedByMinutes: 2,
		Metadata:         json.RawMessage(`{"tenant_id": "tenant-123"}`),
		CreatedAt:        time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:        time.Now(),
	}
}

// GenerateMockDataSources creates multiple mock data sources
func GenerateMockDataSources(orgID uuid.UUID, count int) []models.DataSource {
	sources := make([]models.DataSource, count)
	sourceTypes := []struct {
		Type     string
		Name     string
		Provider string
	}{
		{"microsoft_365", "Microsoft 365", "microsoft"},
		{"azure", "Azure AD", "microsoft"},
		{"aws", "AWS CloudTrail", "aws"},
		{"google_workspace", "Google Workspace", "google"},
	}

	for i := 0; i < count; i++ {
		source := GenerateMockDataSource(orgID)
		st := sourceTypes[i%len(sourceTypes)]
		source.Type = st.Type
		source.Name = st.Name
		source.Provider = strPtr(st.Provider)
		sources[i] = *source
	}
	return sources
}

// GenerateMockParser creates a mock parser for testing
func GenerateMockParser(orgID uuid.UUID) *models.Parser {
	id := uuid.New()
	sourceID := uuid.New()

	return &models.Parser{
		ID:              id,
		OrganizationID:  orgID,
		SourceID:        &sourceID,
		Name:            "Microsoft 365 Login Parser",
		Description:     strPtr("Parses Microsoft 365 user login events"),
		ParserType:      types.ParserTypeJSON,
		Status:          types.ParserStatusActive,
		Tags:            []string{"authentication", "microsoft"},
		Logic:           map[string]interface{}{"extract": map[string]string{"user": "actor_email", "timestamp": "occurred_at"}},
		Mappings:        []map[string]interface{}{{"field": "user", "target": "actor.email"}},
		EventsParsed24h: 5000,
		ErrorRate:       0.02,
		CreatedAt:       time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt:       time.Now(),
	}
}

// GenerateMockParsers creates multiple mock parsers
func GenerateMockParsers(orgID uuid.UUID, count int) []models.Parser {
	parsers := make([]models.Parser, count)
	parserTypes := []types.ParserType{types.ParserTypeJSON, types.ParserTypeRegex, types.ParserTypeCSV}

	for i := 0; i < count; i++ {
		parser := GenerateMockParser(orgID)
		parser.ParserType = parserTypes[i%len(parserTypes)]
		parser.Name = parser.Name + " " + string(rune('A'+i%3))
		parsers[i] = *parser
	}
	return parsers
}
