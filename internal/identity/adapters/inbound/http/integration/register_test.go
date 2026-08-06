package integration

import (
	"net/http"
	"testing"

	"sage-backend/internal/identity/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterUser_Integration(t *testing.T) {
	t.Run("should register user and organization successfully", func(t *testing.T) {
		harness := setUpUsersApp(t)

		// 1. Get industry ID from database
		var industryID string
		err := harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)
		require.NoError(t, err, "Industries table must contain seeded rows")

		reqBody := dto.OnboardingRequest{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane.doe@example.com",
			Password:    "SecurePassword123!",
			CompanyName: "Acme Corp",
			IndustryId:  industryID,
		}

		resp, err := performRequest(
			harness.App,
			http.MethodPost,
			"/test/v1/auth/register",
			reqBody,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated, "Expected 200 or 201, got %d", resp.StatusCode)

		res := decodeResponse(t, resp)
		assert.Equal(t, "User and organization created successfully. Please verify your email.", res.Message)

		// Verify database insertion in sage_db_test PostgreSQL
		var userCount int
		err = harness.Database.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "jane.doe@example.com").Scan(&userCount)
		require.NoError(t, err)
		assert.Equal(t, 1, userCount)

		var orgCount int
		err = harness.Database.QueryRow("SELECT COUNT(*) FROM organizations WHERE name = $1", "Acme Corp").Scan(&orgCount)
		require.NoError(t, err)
		assert.Equal(t, 1, orgCount)
	})

	t.Run("should reject invalid email format", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)

		reqBody := dto.OnboardingRequest{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "invalid-email",
			Password:    "SecurePassword123!",
			CompanyName: "Acme Corp",
			IndustryId:  industryID,
		}

		resp, err := performRequest(
			harness.App,
			http.MethodPost,
			"/test/v1/auth/register",
			reqBody,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("should reject duplicate email", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)

		reqBody := dto.OnboardingRequest{
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane.doe@example.com",
			Password:    "SecurePassword123!",
			CompanyName: "Acme Corp",
			IndustryId:  industryID,
		}

		// First registration - should succeed
		resp1, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", reqBody)
		require.NoError(t, err)
		resp1.Body.Close()
		require.True(t, resp1.StatusCode == http.StatusOK || resp1.StatusCode == http.StatusCreated)

		// Duplicate registration - should fail
		resp2, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", reqBody)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.True(t, resp2.StatusCode >= 400, "Duplicate registration should return error status code")
	})
}
