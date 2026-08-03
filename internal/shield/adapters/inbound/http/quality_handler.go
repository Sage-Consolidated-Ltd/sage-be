package http

import (
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/usecase/dto"
	"sage-backend/internal/shield/ports/inbound"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QualityHandler struct {
	service inbound.DataQualityUseCase
}

func NewQualityHandler(service inbound.DataQualityUseCase) *QualityHandler {
	return &QualityHandler{service: service}
}

func (h *QualityHandler) GetDataQualitySummary(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	scan, err := h.service.GetSummary(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_SUMMARY", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Quality summary", scan.ToResponse())
}
func (h *QualityHandler) RunDataQualityScan(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	result, err := h.service.RunScan(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_START_SCAN", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Scan started", result)
}
func (h *QualityHandler) GetDataQualityBreakdown(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	sourceIDStr := c.Query("source_id")
	var srcID *uuid.UUID
	if sourceIDStr != "" {
		if uid, err := uuid.Parse(sourceIDStr); err == nil {
			srcID = &uid
		}
	}
	metrics, err := h.service.GetBreakdown(c.Context(), orgID, srcID, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_BREAKDOWN", err.Error())
	}
	items := make([]interface{}, len(metrics))
	for i, m := range metrics {
		// enrich with source name? For now, use minimal response
		items[i] = m.ToResponse("") // source name empty; handler could fetch source if needed
	}
	paginated := map[string]interface{}{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Quality breakdown", paginated)
}
func (h *QualityHandler) GetAIAnalysis(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	insights, err := h.service.GetAIAnalysis(c.Context(), orgID, nil)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_AI_ANALYSIS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "AI analysis", map[string]interface{}{"insights": insights})
}
func (h *QualityHandler) ApplySuggestedFix(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	var req dto.ApplySuggestedFixRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	suggestionUUID, err := uuid.Parse(req.SuggestionID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_SUGGESTION_ID", err.Error())
	}
	if err := h.service.ApplySuggestedFix(c.Context(), orgID, suggestionUUID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_APPLY_FIX", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Fix applied", nil)
}
func (h *QualityHandler) GetSuggestedFixDiff(c *fiber.Ctx) error {
	suggestionIDStr := c.Query("suggestion_id")
	parserIDStr := c.Query("parser_id")
	if suggestionIDStr == "" || parserIDStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "MISSING_PARAMETERS", "suggestion_id and parser_id required")
	}
	suggestionID, err := uuid.Parse(suggestionIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_SUGGESTION_ID", err.Error())
	}
	parserID, err := uuid.Parse(parserIDStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PARSER_ID", err.Error())
	}
	diff, err := h.service.GetSuggestedFixDiff(c.Context(), suggestionID, parserID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_DIFF", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Diff preview", diff)
}
func (h *QualityHandler) DownloadDataQualityReport(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	format := c.Query("format", "csv")
	if format != "csv" && format != "pdf" && format != "json" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_FORMAT", "Format must be csv, pdf, or json")
	}
	var startTime, endTime *time.Time
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			startTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			endTime = &t
		}
	}
	data, filename, err := h.service.DownloadDataQualityReport(c.Context(), orgID, format, startTime, endTime)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GENERATE_REPORT", err.Error())
	}
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.Send(data)
}
