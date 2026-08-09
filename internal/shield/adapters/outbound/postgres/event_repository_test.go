package postgres

import (
	"context"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityEventRepository_StructScanMapping(t *testing.T) {
	// Skip if test database is not available
	testDBURL := "postgres://peterpaul:sage_dev_password@localhost:5432/sage_db_test?sslmode=disable"
	database, err := db.ConnectDB(&config.BaseConfig{
		DatabaseUrl:    testDBURL,
		DBMAXOpenConns: 5,
		DBMAXIdleConns: 5,
		DBConnMAXLife:  300,
	})
	if err != nil {
		t.Skip("Database not available for integration testing:", err)
		return
	}
	defer database.Close()

	repo := NewSecurityEventRepository(database)
	ctx := context.Background()

	orgID := uuid.New()
	sourceID := uuid.New()
	userID := uuid.New()

	// Seed test user, org, and data source
	_, err = database.ExecContext(ctx, "INSERT INTO users (id, first_name, last_name, email, password_hash) VALUES ($1, 'Test', 'User', $2, 'hash')", userID, "test-"+userID.String()+"@example.com")
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "INSERT INTO organizations (id, name, owner_id) VALUES ($1, 'Test Org', $2)", orgID, userID)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "INSERT INTO data_sources (id, organization_id, name, provider, type, status) VALUES ($1, $2, 'Test Source', 'okta', 'api', 'active')", sourceID, orgID)
	require.NoError(t, err)

	sev := types.SeverityHigh
	evt := &domain.SecurityEvent{
		OrganizationID: orgID,
		SourceID:       sourceID,
		Source:         "okta",
		EventType:      "user.login.failed",
		EventCategory:  "authentication",
		Severity:       &sev,
		ParseStatus:    types.ParseStatusPending,
		RawPayload:     map[string]interface{}{"event": "test"},
		OccurredAt:     time.Now(),
	}

	err = repo.CreateEvent(ctx, evt)
	require.NoError(t, err)

	// Test SearchEvents row deserialization into SecurityEvent struct
	filters := map[string]interface{}{"source_id": sourceID.String()}
	events, total, err := repo.SearchEvents(ctx, orgID, filters, 1, 10)
	require.NoError(t, err, "SearchEvents MUST NOT fail with struct mapping errors")
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, events)
	assert.Equal(t, orgID, events[0].OrganizationID)
}
