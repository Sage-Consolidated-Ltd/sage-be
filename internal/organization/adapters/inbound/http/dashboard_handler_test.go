package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sage-backend/internal/organization/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDashboardUseCase struct {
	getDashboardFn     func(ctx context.Context, orgID uuid.UUID, tab domain.DashboardTab) (*domain.OrganizationDashboard, error)
	refreshDashboardFn func(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error)
}

func (m *mockDashboardUseCase) GetDashboard(ctx context.Context, orgID uuid.UUID, tab domain.DashboardTab) (*domain.OrganizationDashboard, error) {
	if m.getDashboardFn != nil {
		return m.getDashboardFn(ctx, orgID, tab)
	}
	return nil, nil
}

func (m *mockDashboardUseCase) RefreshDashboard(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
	if m.refreshDashboardFn != nil {
		return m.refreshDashboardFn(ctx, orgID)
	}
	return nil, nil
}

func TestDashboardHandler_GetDashboard_Unauthorized(t *testing.T) {
	app := fiber.New()
	handler := NewDashboardHandler(&mockDashboardUseCase{}, NewDashboardHub())
	app.Get("/dashboard", handler.GetDashboard)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDashboardHandler_GetDashboard_Success(t *testing.T) {
	app := fiber.New()
	testOrgID := uuid.New()

	mockUC := &mockDashboardUseCase{
		getDashboardFn: func(ctx context.Context, orgID uuid.UUID, tab domain.DashboardTab) (*domain.OrganizationDashboard, error) {
			assert.Equal(t, testOrgID, orgID)
			assert.Equal(t, domain.TabOverview, tab)
			return &domain.OrganizationDashboard{
				OrganizationID: testOrgID,
				Tab:            tab,
				SecurityScore: &domain.SecurityScore{
					OverallScore: 94,
				},
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}

	handler := NewDashboardHandler(mockUC, NewDashboardHub())

	// Middleware injects orgID into locals to emulate active session
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("orgID", testOrgID.String())
		return c.Next()
	})
	app.Get("/dashboard", handler.GetDashboard)

	req := httptest.NewRequest(http.MethodGet, "/dashboard?tab=overview", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	require.NoError(t, err)
	assert.True(t, res["success"].(bool))
	assert.NotNil(t, res["data"])
}

func TestDashboardHandler_RefreshDashboard_Success(t *testing.T) {
	app := fiber.New()
	testOrgID := uuid.New()

	mockUC := &mockDashboardUseCase{
		refreshDashboardFn: func(ctx context.Context, orgID uuid.UUID) (*domain.OrganizationDashboard, error) {
			assert.Equal(t, testOrgID, orgID)
			return &domain.OrganizationDashboard{
				OrganizationID: testOrgID,
				SecurityScore: &domain.SecurityScore{
					OverallScore: 95,
				},
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}

	handler := NewDashboardHandler(mockUC, NewDashboardHub())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("orgID", testOrgID.String())
		return c.Next()
	})
	app.Post("/dashboard/refresh", handler.RefreshDashboard)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/refresh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	require.NoError(t, err)
	assert.True(t, res["success"].(bool))
}

func TestDashboardHandler_UpgradeWebSocket_RequiresUpgrade(t *testing.T) {
	app := fiber.New()
	handler := NewDashboardHandler(&mockDashboardUseCase{}, NewDashboardHub())
	app.Get("/dashboard/ws", handler.UpgradeWebSocket)

	// Standard GET without WebSocket upgrade headers must fail with 426
	req := httptest.NewRequest(http.MethodGet, "/dashboard/ws", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUpgradeRequired, resp.StatusCode)
}

func TestDashboardHub_Lifecycle(t *testing.T) {
	hub := NewDashboardHub()
	require.NotNil(t, hub)

	orgID := uuid.New().String()

	// Broadcast on empty hub does not crash
	assert.NotPanics(t, func() {
		hub.Broadcast(orgID, map[string]string{"status": "ok"})
	})
}
