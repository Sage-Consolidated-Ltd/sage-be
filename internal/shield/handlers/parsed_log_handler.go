package handlers

import (
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ParsedLogHandler struct {
	logService services.ParsedLogService
}

func NewParsedLogHandler(logService services.ParsedLogService) *ParsedLogHandler {
	return &ParsedLogHandler{
		logService: logService,
	}
}

func (h *ParsedLogHandler) SearchParsedLogs(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}

	params := models.SearchParams{
		OrganizationID: orgID,
		Limit:          50,
	}

	if q := c.Query("query"); q != "" {
		// AST path — single Splunk-style query string
		ast := h.logService.ParseQueryString(q)
		astParams, err := h.logService.ASTToSearchParams(ast)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		astParams.OrganizationID = orgID
		astParams.Limit = params.Limit
		params = astParams
	} else {
		// Flat-param fallback — existing behavior, unchanged
		if v := c.Query("level"); v != "" {
			params.Level = &v
		}
		if v := c.Query("q"); v != "" {
			params.FreeText = &v
		}
		if v := c.Query("source"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				return response.Error(c, fiber.StatusBadRequest, "invalid source id", nil)
			}
			params.DataSourceID = &id
		}
	}

	// Time range, cursor, and limit apply regardless of which path was used above
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid from timestamp", nil)
		}
		params.From = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid to timestamp", nil)
		}
		params.To = &t
	}
	if v := c.Query("cursor"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid cursor", nil)
		}
		params.Cursor = &t
	}
	if v := c.QueryInt("limit"); v > 0 {
		params.Limit = v
	}

	result, err := h.logService.SearchLogs(c.Context(), params)
	if err != nil {
		logger.Error("Error searching parsed logs at ParsedLogHandler.SearchParsedLogs", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "search failed", err)
	}

	return response.JSON(c, fiber.StatusOK, "Logs retrieved successfully", result)
}