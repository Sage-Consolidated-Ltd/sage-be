package http

import (
	"sage-backend/internal/organization/domain"
	_ "sage-backend/internal/shared/response"
)

type OrganizationDashboardResponse struct {
	Success bool                         `json:"success"`
	Message string                       `json:"message"`
	Data    domain.OrganizationDashboard `json:"data"`
}

// @Summary Get Organization Dashboard Snapshot
// @Description Retrieve the unified, materialized security & operational dashboard snapshot for the organization. Returns all 10 widgets in a single optimized payload. Supports pre-filtering by tab (overview, assets, health, identity).
// @Tags Organization Dashboard
// @Accept json
// @Produce json
// @Param tab query string false "Dashboard Tab" Enums(overview, assets, health, identity) default(overview)
// @Success 200 {object} OrganizationDashboardResponse "Unified dashboard snapshot"
// @Failure 401 {object} response.ErrorResponse "Unauthorized: Active session or organization required"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /organization/dashboard [get]
func _GetDashboard() {}

// @Summary Refresh Organization Dashboard Snapshot
// @Description Forces a fresh OLTP calculation across threats, security events, data sources, members, and alerts, warms the Redis cache, and broadcasts the updated snapshot to all connected WebSocket clients.
// @Tags Organization Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} OrganizationDashboardResponse "Recalculated fresh dashboard snapshot"
// @Failure 401 {object} response.ErrorResponse "Unauthorized: Active session required"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /organization/dashboard/refresh [post]
func _RefreshDashboard() {}

// @Summary Stream Organization Dashboard Real-Time Updates
// @Description Connect via WebSocket (ws:// or wss://) to receive live streaming updates for the organization's dashboard whenever fresh security events, threat logs, or manual refreshes occur.
// @Tags Organization Dashboard
// @Param org_id query string false "Organization UUID (if cookie authentication is not available)"
// @Success 101 "Switching Protocols to WebSocket"
// @Failure 401 {object} response.ErrorResponse "Unauthorized: Organization ID required"
// @Failure 426 {string} string "Upgrade Required: Missing WebSocket headers"
// @Router /organization/dashboard/ws [get]
func _DashboardWebSocket() {}
