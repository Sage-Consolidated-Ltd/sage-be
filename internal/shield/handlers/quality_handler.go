package handlers

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QualityHandler struct {
	service services.DataQualityServiceInt
}

func NewQualityHandler(service services.DataQualityServiceInt) *QualityHandler {
	return &QualityHandler{service: service}
}

// @Summary Get Data Quality Summary
// @Description Returns overall data quality metrics including score, parsing errors, missing fields, unmapped logs, and duplicate events.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality [get]
func (h *QualityHandler) GetDataQualitySummary(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	scan, err := h.service.GetSummary(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_SUMMARY", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Quality summary", scan.ToResponse())
}

// @Summary Run Data Quality Scan
// @Description Queues a data quality scan to detect parsing errors, missing fields, duplicates, and unmapped logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality/scan [post]
func (h *QualityHandler) RunDataQualityScan(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	result, err := h.service.RunScan(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_START_SCAN", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Scan started", result)
}

// @Summary Get Data Quality Breakdown
// @Description Returns data quality metrics per data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Quality status"
// @Param source_id query string false "Data source ID"
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality/breakdown [get]
func (h *QualityHandler) GetDataQualityBreakdown(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
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

// @Summary Get Data Quality AI Analysis
// @Description Returns AI-generated analysis and suggested fixes for data quality issues.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality/ai-analysis [get]
func (h *QualityHandler) GetAIAnalysis(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	insights, err := h.service.GetAIAnalysis(c.Context(), orgID, nil)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_AI_ANALYSIS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "AI analysis", map[string]interface{}{"insights": insights})
}

// @Summary Apply Suggested Data Quality Fix
// @Description Applies an AI-suggested parser or field mapping fix to improve data quality.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param request body requests.ApplySuggestedFixRequest true "Apply Suggested Fix Request"
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality/apply-suggested-fix [post]
func (h *QualityHandler) ApplySuggestedFix(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	var req requests.ApplySuggestedFixRequest
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

// @Summary Preview Suggested Fix Diff
// @Description Returns before and after diff for an AI-suggested parser or mapping fix.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param suggestion_id query string true "Suggestion ID"
// @Param parser_id query string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /logs-data/data-quality/diff [get]
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

// @Summary Download Data Quality Report
// @Description Downloads data quality report in csv, pdf, or json format.
// @Tags Logs & Data
// @Accept json
// @Produce application/octet-stream
// @Param format query string false "Report format"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {file} file
// @Router /logs-data/data-quality/report [get]
func (h *QualityHandler) DownloadDataQualityReport(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
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
