package integration

import (
	"context"
	"net/http"
	"testing"

	"sage-backend/internal/identity/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCurrentUser_Integration(t *testing.T) {
	t.Run("should fetch profile of authenticated user", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)

		// 1. Register
		regPayload := dto.OnboardingRequest{
			FirstName:   "Profile",
			LastName:    "Tester",
			Email:       "profile.tester@example.com",
			Password:    "SecurePassword123!",
			CompanyName: "Profile Corp",
			IndustryId:  industryID,
		}
		regResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", regPayload)
		require.NoError(t, err)
		regResp.Body.Close()

		// 2. Login & extract cookie
		loginPayload := dto.LoginRequest{
			Email:    "profile.tester@example.com",
			Password: "SecurePassword123!",
		}
		loginResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", loginPayload)
		require.NoError(t, err)
		cookieHeader := extractCookieHeader(loginResp)
		loginResp.Body.Close()
		require.NotEmpty(t, cookieHeader)

		// 3. Request profile with cookie
		resp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/profile/me", nil, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Profile retrieved successfully", res.Message)

		// 4. Update contact identity fields (phone_number & backup_email)
		updateIdentityReq := dto.UpdateIdentityRequest{
			PhoneNumber: "+447709432112",
			BackupEmail: "backup.tester@example.com",
		}
		upResp, err := performRequestWithCookie(harness.App, http.MethodPatch, "/test/v1/profile/me", updateIdentityReq, cookieHeader)
		require.NoError(t, err)
		upResp.Body.Close()
		require.Equal(t, http.StatusOK, upResp.StatusCode)

		// 5. Update preferences (date_format)
		updatePrefsReq := dto.UpdatePreferencesRequest{
			Theme:      "dark",
			Timezone:   "Europe/London",
			Language:   "en",
			DateFormat: "DD/MM/YYYY",
		}
		prefsResp, err := performRequestWithCookie(harness.App, http.MethodPatch, "/test/v1/profile/preferences", updatePrefsReq, cookieHeader)
		require.NoError(t, err)
		prefsResp.Body.Close()
		require.Equal(t, http.StatusOK, prefsResp.StatusCode)

		// 6. Update notifications (product_updates & weekly_summary)
		updateNotifReq := dto.UpdateNotificationsRequest{
			EmailEnabled:           true,
			ProductUpdates:         true,
			WeeklySummary:          true,
			AlertSeverityThreshold: "medium",
		}
		notifResp, err := performRequestWithCookie(harness.App, http.MethodPatch, "/test/v1/profile/notifications", updateNotifReq, cookieHeader)
		require.NoError(t, err)
		notifResp.Body.Close()
		require.Equal(t, http.StatusOK, notifResp.StatusCode)

		// 7. Verify session storage in Redis DB 2
		ctx := context.Background()
		keys, err := harness.Redis.Keys(ctx, "*").Result()
		require.NoError(t, err)
		assert.NotEmpty(t, keys, "Redis DB 2 must contain active session keys")
	})

	t.Run("should reject unauthenticated profile request", func(t *testing.T) {
		harness := setUpUsersApp(t)

		resp, err := performRequest(harness.App, http.MethodGet, "/test/v1/profile/me", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
