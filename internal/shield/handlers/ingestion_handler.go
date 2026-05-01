package handlers

import (
	"time"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LogsDataHandler struct {
	service services.LogsDataServiceInt
}

func NewLogsDataHandlerWithService(svc services.LogsDataServiceInt) *LogsDataHandler {
	return &LogsDataHandler{service: svc}
}

// Helper to extract orgID from session
func getSessionInfo(c *fiber.Ctx) (orgID uuid.UUID, userID string, role string, ok bool) {
	// Fallback to locals for testing if store is nil
	if config.Store == nil {
		if orgIDVal := c.Locals("orgID"); orgIDVal != nil {
			if orgUUID, ok := orgIDVal.(uuid.UUID); ok {
				return orgUUID, "", "", true
			}
		}
		return uuid.Nil, "", "", false
	}
	sess, err := config.Store.Get(c)
	if err != nil {
		return uuid.Nil, "", "", false
	}
	orgStr, _ := sess.Get("organizationID").(string)
	uid, _ := sess.Get("userID").(string)
	r, _ := sess.Get("role").(string)
	if orgStr == "" {
		return uuid.Nil, "", "", false
	}
	orgUUID, err := uuid.Parse(orgStr)
	if err != nil {
		return uuid.Nil, "", "", false
	}
	return orgUUID, uid, r, true
}

// @Summary Get Ingestion Health Summary
// @Description Returns ingestion health metrics including total events ingested, active sources, delayed sources, and ingestion errors.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/ingestion-health [get]
func (h *LogsDataHandler) GetIngestionHealth(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	health, err := h.service.GetIngestionHealth(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_HEALTH", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Ingestion health retrieved", health)
}

// @Summary Refresh Ingestion Health
// @Description Queues a background job to refresh ingestion health metrics.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/ingestion-health/refresh [post]
func (h *LogsDataHandler) RefreshIngestionHealth(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	result, err := h.service.RefreshIngestionHealth(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_QUEUE_REFRESH", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Refresh queued", result)
}

// @Summary List Data Sources
// @Description Returns connected data sources with ingestion status, events today, last event time, and health status.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by source name or description"
// @Param type query string false "Source type"
// @Param status query string false "Source status"
// @Success 200 {object} response.Response
// @Router /logs-data/sources [get]
func (h *LogsDataHandler) ListSources(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	filters := map[string]interface{}{
		"status": c.Query("status"),
		"type":   c.Query("type"),
		"search": c.Query("search"),
	}
	sources, total, err := h.service.ListSources(c.Context(), orgID, filters, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_LIST_SOURCES", err.Error())
	}
	items := make([]interface{}, len(sources))
	for i, src := range sources {
		items[i] = src.ToResponse()
	}
	paginated := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Sources retrieved", paginated)
}

// @Summary Get Data Source Detail
// @Description Returns detailed health and metadata for a single data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /logs-data/sources/{id} [get]
func (h *LogsDataHandler) GetSource(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	source, err := h.service.GetSource(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "SOURCE_NOT_FOUND", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Source retrieved", source.ToResponse())
}

// @Summary Sync Data Source
// @Description Manually queues a sync job for a data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /logs-data/sources/{id}/sync [post]
func (h *LogsDataHandler) SyncSource(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	result, err := h.service.SyncSource(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_QUEUE_SYNC", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Sync queued", result)
}

// @Summary Disconnect Data Source
// @Description Disconnects a data source and stops future ingestion while preserving historical logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /logs-data/sources/{id}/disconnect [post]
func (h *LogsDataHandler) DisconnectSource(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	if err := h.service.DisconnectSource(c.Context(), id, orgID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_DISCONNECT", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Source disconnected", nil)
}

// @Summary View Source Logs
// @Description Returns logs ingested from a specific data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Data Source ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param event_type query string false "Event type"
// @Param severity query string false "Severity"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {object} response.Response
// @Router /logs-data/sources/{id}/logs [get]
func (h *LogsDataHandler) GetSourceLogs(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	sourceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	filters := map[string]interface{}{
		"event_type": c.Query("event_type"),
		"severity":   c.Query("severity"),
		"start_time": c.Query("start_time"),
		"end_time":   c.Query("end_time"),
	}
	logs, total, err := h.service.GetSourceLogs(c.Context(), sourceID, orgID, filters, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_LOGS", err.Error())
	}
	items := make([]interface{}, len(logs))
	for i, log := range logs {
		items[i] = log.ToResponse()
	}
	paginated := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Logs retrieved", paginated)
}

// @Summary Get Ingestion Volume Over Time
// @Description Returns time-series ingestion volume for charts, optionally filtered by source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Param interval query string false "Aggregation interval"
// @Param source_id query string false "Data source ID"
// @Success 200 {object} response.Response
// @Router /logs-data/ingestion-health/volume [get]
func (h *LogsDataHandler) GetIngestionVolume(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	interval := c.Query("interval", "hour")
	var srcID *uuid.UUID
	if s := c.Query("source_id"); s != "" {
		if uid, err := uuid.Parse(s); err == nil {
			srcID = &uid
		}
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
	volume, err := h.service.GetIngestionVolume(c.Context(), orgID, startTime, endTime, interval, srcID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_VOLUME", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Volume retrieved", volume)
}

// @Summary Get Ingestion Notifications
// @Description Returns ingestion warnings and AI insights related to source delays, drops, parsing errors, and schema changes.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/ingestion-health/notifications [get]
func (h *LogsDataHandler) GetIngestionNotifications(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	notifs, err := h.service.GetIngestionNotifications(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_NOTIFICATIONS", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Notifications retrieved", notifs)
}

// @Summary Download Ingestion Health Report
// @Description Downloads ingestion health report in csv, pdf, or json format.
// @Tags Logs & Data
// @Accept json
// @Produce application/octet-stream
// @Param format query string false "Report format"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {file} file
// @Router /logs-data/ingestion-health/report [get]
func (h *LogsDataHandler) DownloadIngestionHealthReport(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
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
	data, filename, err := h.service.DownloadIngestionHealthReport(c.Context(), orgID, format, startTime, endTime)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GENERATE_REPORT", err.Error())
	}
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.Send(data)
}
