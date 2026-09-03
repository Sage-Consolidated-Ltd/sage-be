package http

import (
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/ports/inbound"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	service inbound.DashboardUseCase
}

func NewDashboardHandler(service inbound.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// Widget 1: Security Posture Score
func (h *DashboardHandler) GetSecurityPostureScore(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetSecurityPostureScore(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_POSTURE_SCORE", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Security posture score retrieved", res)
}

// Widget 3: Identity & Access Health
func (h *DashboardHandler) GetIdentityHealthSummary(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetIdentityHealthSummary(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_IDENTITY_HEALTH", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Identity health summary retrieved", res)
}

// Widget 4: Endpoint Protection Coverage
func (h *DashboardHandler) GetAssetProtectionCoverage(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetAssetProtectionCoverage(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_ASSET_COVERAGE", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Asset protection coverage retrieved", res)
}

// Widget 5: Threat Intelligence Feeds
func (h *DashboardHandler) GetThreatIntelFeedsSummary(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetThreatIntelFeedsSummary(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_THREAT_INTEL", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Threat intelligence feeds summary retrieved", res)
}

// Widget 6: Active Incidents Table
func (h *DashboardHandler) GetActiveIncidents(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	limit := c.QueryInt("limit", 10)
	res, err := h.service.GetActiveIncidents(c.Context(), orgID, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_ACTIVE_INCIDENTS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Active incidents retrieved", res)
}

// Widget 7: Dangerous Threats Risk Distribution
func (h *DashboardHandler) GetAssetRiskDistribution(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetAssetRiskDistribution(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_ASSET_RISK", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Asset risk distribution retrieved", res)
}

// Widget 8: Compliance Risk Indicators
func (h *DashboardHandler) GetComplianceRiskIndicators(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetComplianceRiskIndicators(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_COMPLIANCE_INDICATORS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Compliance risk indicators retrieved", res)
}

// Widget 9: Threat Severity Trends Line Chart
func (h *DashboardHandler) GetThreatTrends(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	currentMonthQuery := c.Query("current_month", c.Query("month", ""))
	prevMonthQuery := c.Query("previous_month", c.Query("prev_month", c.Query("compare_month", "")))

	res, err := h.service.GetThreatTrends(c.Context(), orgID, currentMonthQuery, prevMonthQuery)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_THREAT_TRENDS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Threat severity trends retrieved", res)
}

// Widget 10: Live Geo Threat Origins Map
func (h *DashboardHandler) GetGeoThreats(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	res, err := h.service.GetGeoThreats(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_GEO_THREATS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Geo threat origins retrieved", res)
}
