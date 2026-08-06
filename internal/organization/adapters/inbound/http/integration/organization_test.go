package integration

import (
	"net/http"
	"testing"

	identity_dto "sage-backend/internal/identity/usecase/dto"
	org_dto "sage-backend/internal/organization/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicCompanyRoutes_Integration(t *testing.T) {
	t.Run("should fetch industries without authentication", func(t *testing.T) {
		harness := setUpOrgApp(t)

		resp, err := performRequest(harness.App, http.MethodGet, "/test/v1/company/industries", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Industries retrieved successfully", res.Message)
		assert.NotEmpty(t, res.Data)
	})

	t.Run("should fetch organization roles without authentication", func(t *testing.T) {
		harness := setUpOrgApp(t)

		resp, err := performRequest(harness.App, http.MethodGet, "/test/v1/company/organization-roles", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Organization roles retrieved successfully", res.Message)
		assert.NotEmpty(t, res.Data)
	})
}

func TestOrganizationProtectedRoutes_Integration(t *testing.T) {
	harness := setUpOrgApp(t)

	var industryID string
	_ = harness.Database.QueryRow("SELECT id FROM industries LIMIT 1").Scan(&industryID)
	if industryID == "" {
		_ = harness.Database.QueryRow("INSERT INTO industries (name) VALUES ('Software') ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&industryID)
	}
	require.NotEmpty(t, industryID)

	// 1. Onboard owner user
	regPayload := identity_dto.OnboardingRequest{
		FirstName:   "OrgOwner",
		LastName:    "Tester",
		Email:       "org.owner@example.com",
		Password:    "SecurePass123!",
		CompanyName: "Acme Org Corp",
		IndustryId:  industryID,
	}
	regResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/register", regPayload)
	require.NoError(t, err)
	regResp.Body.Close()

	// 2. Login to get authenticated session cookie
	loginPayload := identity_dto.LoginRequest{
		Email:    "org.owner@example.com",
		Password: "SecurePass123!",
	}
	loginResp, err := performRequest(harness.App, http.MethodPost, "/test/v1/auth/login", loginPayload)
	require.NoError(t, err)
	cookieHeader := extractCookieHeader(loginResp)
	loginResp.Body.Close()
	require.NotEmpty(t, cookieHeader)

	t.Run("should fetch organization profile", func(t *testing.T) {
		resp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/", nil, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Organization retrieved successfully", res.Message)
	})

	t.Run("should fetch aggregated organization profile details (company, environment, branding)", func(t *testing.T) {
		resp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/profile", nil, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Organization profile retrieved successfully", res.Message)
	})

	t.Run("should update company details section", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":                   "Updated Cyber Defense Ltd",
			"industry":               "Financial Services",
			"primary_contact_email": "admin@acmecorp.com",
			"support_email":         "support@acmecorp.com",
		}
		resp, err := performRequestWithCookie(harness.App, http.MethodPatch, "/test/v1/organization/profile/details", payload, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Company details updated successfully", res.Message)
	})

	t.Run("should update organization branding section", func(t *testing.T) {
		payload := map[string]interface{}{
			"logo_light_url":  "https://example.com/logo-light.png",
			"logo_dark_url":   "https://example.com/logo-dark.png",
			"show_in_reports": true,
		}
		resp, err := performRequestWithCookie(harness.App, http.MethodPatch, "/test/v1/organization/branding", payload, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Organization branding updated successfully", res.Message)
	})

	t.Run("should fetch organization members", func(t *testing.T) {
		resp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/members", nil, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Members retrieved successfully", res.Message)
	})

	t.Run("should fetch organization settings", func(t *testing.T) {
		resp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/settings", nil, cookieHeader)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		res := decodeResponse(t, resp)
		assert.Equal(t, "Settings retrieved successfully", res.Message)
	})

	t.Run("should fetch permissions and permission groups", func(t *testing.T) {
		// List permissions
		pResp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/permissions", nil, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, pResp.StatusCode)
		pRes := decodeResponse(t, pResp)
		assert.Equal(t, "Permissions retrieved successfully", pRes.Message)
		pResp.Body.Close()

		// List permission groups
		pgResp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/permission-groups", nil, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, pgResp.StatusCode)
		pgRes := decodeResponse(t, pgResp)
		assert.Equal(t, "Permission groups retrieved successfully", pgRes.Message)
		pgResp.Body.Close()
	})

	t.Run("should perform CRUD on custom roles", func(t *testing.T) {
		// 1. Create custom role
		createPayload := org_dto.CreateCustomRoleRequest{
			Name:             "Security Analyst",
			Description:      "Custom role for analyzing security alerts",
			Permissions:      []string{},
			PermissionGroups: []string{},
		}
		createResp, err := performRequestWithCookie(harness.App, http.MethodPost, "/test/v1/organization/custom-roles", createPayload, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, createResp.StatusCode)
		createRes := decodeResponse(t, createResp)
		assert.Equal(t, "Custom role created successfully", createRes.Message)
		createResp.Body.Close()

		// 2. List custom roles
		listResp, err := performRequestWithCookie(harness.App, http.MethodGet, "/test/v1/organization/custom-roles", nil, cookieHeader)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, listResp.StatusCode)
		listRes := decodeResponse(t, listResp)
		assert.Equal(t, "Custom roles retrieved successfully", listRes.Message)
		listResp.Body.Close()
	})
}

func TestOrganizationUnauthenticatedAccess_Integration(t *testing.T) {
	harness := setUpOrgApp(t)

	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/test/v1/organization/"},
		{http.MethodGet, "/test/v1/organization/members"},
		{http.MethodGet, "/test/v1/organization/settings"},
		{http.MethodGet, "/test/v1/organization/permissions"},
		{http.MethodGet, "/test/v1/organization/permission-groups"},
		{http.MethodGet, "/test/v1/organization/custom-roles"},
	}

	for _, ep := range protectedEndpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			resp, err := performRequest(harness.App, ep.method, ep.path, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}
