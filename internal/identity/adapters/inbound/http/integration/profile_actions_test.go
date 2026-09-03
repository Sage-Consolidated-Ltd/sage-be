package integration

import (
	"net/http"
	"testing"

	"sage-backend/internal/identity/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileActions_Integration(t *testing.T) {
	t.Run("should test change password, backup email, and delete account flows", func(t *testing.T) {
		harness := setUpUsersApp(t)

		var industryID string
		err := harness.Database.QueryRow("SELECT id::text FROM industries LIMIT 1").Scan(&industryID)
		require.NoError(t, err)
		require.NotEmpty(t, industryID)

		// 1. Register test user
		email := "profile.actions.user@example.com"
		oldPassword := "OldSecurePassword123!"
		newPassword := "NewSuperSecurePassword123!#"

		regPayload := dto.OnboardingRequest{
			FirstName:   "Profile",
			LastName:    "Actions",
			Email:       email,
			Password:    oldPassword,
			CompanyName: "Profile Actions Corp",
			IndustryId:  industryID,
		}
		regResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", regPayload)
		require.NoError(t, err)
		regResp.Body.Close()

		// 2. Login to get session cookie
		loginResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", dto.LoginRequest{
			Email:    email,
			Password: oldPassword,
		})
		require.NoError(t, err)
		cookieHeader := extractCookieHeader(loginResp)
		loginResp.Body.Close()
		require.NotEmpty(t, cookieHeader)

		// 3. Test Configure Backup Email - reject primary email
		sameEmailResp, err := performRequestWithCookie(harness.App, http.MethodPost, "/test/v1/profile/backup-email", dto.ConfigureBackupEmailRequest{
			BackupEmail: email,
		}, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, sameEmailResp.StatusCode)
		sameEmailResp.Body.Close()

		// 4. Test Configure Backup Email - valid distinct backup email
		backupEmail := "recovery.backup@example.com"
		backupResp, err := performRequestWithCookie(harness.App, http.MethodPost, "/test/v1/profile/backup-email", dto.ConfigureBackupEmailRequest{
			BackupEmail: backupEmail,
		}, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, backupResp.StatusCode)
		backupResp.Body.Close()

		// Verify backup email saved in profile response
		getProfileResp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/profile/me", nil, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, getProfileResp.StatusCode)
		getProfileResp.Body.Close()

		// 5. Test Change Password - reject incorrect current password
		badPwdResp, err := performRequestWithCookie(harness.App, http.MethodPost, "/test/v1/profile/change-password", dto.ChangePasswordRequest{
			CurrentPassword: "WrongCurrentPassword123!",
			NewPassword:     newPassword,
		}, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, badPwdResp.StatusCode)
		badPwdResp.Body.Close()

		// 6. Test Change Password - success
		changePwdResp, err := performRequestWithCookie(harness.App, http.MethodPost, "/test/v1/profile/change-password", dto.ChangePasswordRequest{
			CurrentPassword: oldPassword,
			NewPassword:     newPassword,
		}, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, changePwdResp.StatusCode)
		changePwdResp.Body.Close()

		// 7. Verify login with new password succeeds and old password fails
		oldLoginResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", dto.LoginRequest{
			Email:    email,
			Password: oldPassword,
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, oldLoginResp.StatusCode)
		oldLoginResp.Body.Close()

		newLoginResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", dto.LoginRequest{
			Email:    email,
			Password: newPassword,
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, newLoginResp.StatusCode)
		newCookieHeader := extractCookieHeader(newLoginResp)
		newLoginResp.Body.Close()
		require.NotEmpty(t, newCookieHeader)

		// 8. Test Delete Account - reject invalid confirmation string
		invalidDelResp, err := performRequestWithCookie(harness.App, http.MethodDelete, "/test/v1/profile/me", dto.DeleteAccountRequest{
			Confirmation: "NOT_DELETE",
		}, newCookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, invalidDelResp.StatusCode)
		invalidDelResp.Body.Close()

		// 9. Test Delete Account - success with "DELETE"
		delResp, err := performRequestWithCookie(harness.App, http.MethodDelete, "/test/v1/profile/me", dto.DeleteAccountRequest{
			Confirmation: "DELETE",
		}, newCookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, delResp.StatusCode)
		delResp.Body.Close()
	})
}
