package integration

import (
	"net/http"
	"testing"

	"sage-backend/internal/identity/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginUser_Integration(t *testing.T) {
	t.Run("should login successfully with valid credentials", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)

		// Register user first
		regPayload := dto.OnboardingRequest{
			FirstName:   "Login",
			LastName:    "User",
			Email:       "login.user@example.com",
			Password:    "SecurePassword123!",
			CompanyName: "Login Corp",
			IndustryId:  industryID,
		}
		regResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", regPayload)
		require.NoError(t, err)
		regResp.Body.Close()

		// Attempt login
		loginPayload := dto.LoginRequest{
			Email:    "login.user@example.com",
			Password: "SecurePassword123!",
		}

		resp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", loginPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		cookieHeader := extractCookieHeader(resp)
		assert.NotEmpty(t, cookieHeader, "Login response must set session cookie")

		res := decodeResponse(t, resp)
		assert.Equal(t, "Login successful", res.Message)
	})

	t.Run("should reject login with wrong password", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)

		regPayload := dto.OnboardingRequest{
			FirstName:   "Login",
			LastName:    "User",
			Email:       "login.user@example.com",
			Password:    "SecurePassword123!",
			CompanyName: "Login Corp",
			IndustryId:  industryID,
		}
		regResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", regPayload)
		require.NoError(t, err)
		regResp.Body.Close()

		loginPayload := dto.LoginRequest{
			Email:    "login.user@example.com",
			Password: "WrongPassword123!",
		}

		resp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", loginPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400, "Wrong password must return an error status code")
	})

	t.Run("should reject login with non-existent email", func(t *testing.T) {
		harness := setUpUsersApp(t)

		loginPayload := dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "SecurePassword123!",
		}

		resp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", loginPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400, "Non-existent user login must return an error status code")
	})
}
