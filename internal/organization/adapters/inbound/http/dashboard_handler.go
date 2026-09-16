package http

import (
	"context"
	"encoding/json"
	"strings"

	"sage-backend/internal/organization/domain"
	"sage-backend/internal/organization/ports/inbound"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	dashboardServ inbound.DashboardUseCase
	hub           *DashboardHub
}

func NewDashboardHandler(dashboardServ inbound.DashboardUseCase, hub *DashboardHub) *DashboardHandler {
	return &DashboardHandler{
		dashboardServ: dashboardServ,
		hub:           hub,
	}
}

// GetDashboard returns the unified, materialized security & health dashboard snapshot.
// GET /api/v1/organization/dashboard?tab=overview|assets|health|identity
func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	orgIDStr := middlewares.GetOrgIDStr(c)
	if orgIDStr == "" {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized: organization not found in session", nil)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid organization id", nil)
	}

	tabParam := strings.ToLower(strings.TrimSpace(c.Query("tab", "overview")))
	var tab domain.DashboardTab
	switch tabParam {
	case "assets":
		tab = domain.TabAssets
	case "health":
		tab = domain.TabHealth
	case "identity":
		tab = domain.TabIdentity
	default:
		tab = domain.TabOverview
	}

	dashboard, err := h.dashboardServ.GetDashboard(c.Context(), orgID, tab)
	if err != nil {
		logger.Error("Error with DashboardHandler.GetDashboard: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to retrieve dashboard snapshot", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Dashboard snapshot retrieved successfully", dashboard)
}

// RefreshDashboard forces a real-time recalculation of the organization dashboard snapshot and pushes to subscribers.
// POST /api/v1/organization/dashboard/refresh
func (h *DashboardHandler) RefreshDashboard(c *fiber.Ctx) error {
	orgIDStr := middlewares.GetOrgIDStr(c)
	if orgIDStr == "" {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized: organization not found in session", nil)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid organization id", nil)
	}

	freshSnapshot, err := h.dashboardServ.RefreshDashboard(c.Context(), orgID)
	if err != nil {
		logger.Error("Error with DashboardHandler.RefreshDashboard: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to refresh dashboard snapshot", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Dashboard snapshot refreshed and broadcast successfully", freshSnapshot)
}

// UpgradeWebSocket checks for valid WebSocket upgrade headers and extracts the organization ID.
func (h *DashboardHandler) UpgradeWebSocket(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	orgIDStr := middlewares.GetOrgIDStr(c)
	if orgIDStr == "" {
		// Fallback for WebSocket clients passing org_id via query param
		orgIDStr = c.Query("org_id")
	}

	if orgIDStr == "" {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized: organization ID required for live dashboard stream", nil)
	}

	if _, err := uuid.Parse(orgIDStr); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid organization id", nil)
	}

	c.Locals("wsOrgID", orgIDStr)
	return c.Next()
}

// HandleWebSocket manages the live WebSocket connection for streaming real-time dashboard snapshot updates.
// WS /api/v1/organization/dashboard/ws
func (h *DashboardHandler) HandleWebSocket(c *websocket.Conn) {
	orgIDStr, _ := c.Locals("wsOrgID").(string)
	if orgIDStr == "" {
		orgIDStr = c.Query("org_id")
	}

	if orgIDStr == "" {
		_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "missing organization id"))
		_ = c.Close()
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid organization id"))
		_ = c.Close()
		return
	}

	// Register client in the Hub
	h.hub.Register(orgIDStr, c)
	defer h.hub.Unregister(orgIDStr, c)

	// Send initial snapshot upon connection
	if snapshot, err := h.dashboardServ.GetDashboard(context.Background(), orgID, domain.TabOverview); err == nil && snapshot != nil {
		if data, err := json.Marshal(snapshot); err == nil {
			_ = c.WriteMessage(websocket.TextMessage, data)
		}
	}

	// Read incoming messages to maintain connection and support client-triggered refresh commands
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			msgStr := strings.TrimSpace(string(message))
			if msgStr == "refresh" {
				fresh, err := h.dashboardServ.RefreshDashboard(context.Background(), orgID)
				if err == nil && fresh != nil {
					if data, err := json.Marshal(fresh); err == nil {
						_ = c.WriteMessage(websocket.TextMessage, data)
					}
				}
			} else if msgStr == "ping" {
				_ = c.WriteMessage(websocket.TextMessage, []byte(`{"event":"pong"}`))
			}
		}
	}
}
