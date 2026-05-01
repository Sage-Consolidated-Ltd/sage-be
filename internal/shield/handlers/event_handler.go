package handlers

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EventHandler struct {
	service services.LogsServiceInt
}

func NewEventHandler(service services.LogsServiceInt) *EventHandler {
	return &EventHandler{service: service}
}

// @Summary Search Logs
// @Description Searches normalized security events across connected data sources.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param source_id query string false "Data source ID"
// @Param event_type query string false "Event type"
// @Param severity query string false "Severity"
// @Param actor_email query string false "Actor email"
// @Param ip_address query string false "IP address"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Param search query string false "Search text"
// @Success 200 {object} response.Response
// @Router /logs [get]
func (h *EventHandler) SearchLogs(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	// Allow override for testing
	if localOrgID := c.Locals("orgID"); localOrgID != nil {
		if id, ok := localOrgID.(uuid.UUID); ok && id != uuid.Nil {
			orgID = id
		}
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	filters := map[string]interface{}{
		"source_id":      c.Query("source_id"),
		"source":         c.Query("source"),
		"event_type":     c.Query("event_type"),
		"event_category": c.Query("event_category"),
		"severity":       c.Query("severity"),
		"actor_email":    c.Query("actor_email"),
		"ip_address":     c.Query("ip_address"),
		"start_time":     c.Query("start_time"),
		"end_time":       c.Query("end_time"),
		"search":         c.Query("search"),
	}
	logs, total, err := h.service.SearchLogs(c.Context(), orgID, filters, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_SEARCH_LOGS", err.Error())
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

// @Summary Get Log Detail
// @Description Returns full raw and normalized payload for one security event.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Log Event ID"
// @Success 200 {object} response.Response
// @Router /logs/{id} [get]
func (h *EventHandler) GetLogDetail(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	// Allow override for testing
	if localOrgID := c.Locals("orgID"); localOrgID != nil {
		if id, ok := localOrgID.(uuid.UUID); ok && id != uuid.Nil {
			orgID = id
		}
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	event, err := h.service.GetLogByID(c.Context(), orgID, id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "EVENT_NOT_FOUND", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Event retrieved", event.ToResponse())
}

// @Summary Ingest Log Event
// @Description Ingests one raw security event, normalizes it, and stores it for investigation and detection.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param request body requests.IngestLogRequest true "Ingest Log Request"
// @Success 200 {object} response.Response
// @Router /logs/ingest [post]
func (h *EventHandler) IngestLog(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	// Allow override for testing
	if orgID == uuid.Nil {
		if localOrgID := c.Locals("orgID"); localOrgID != nil {
			if id, ok := localOrgID.(uuid.UUID); ok {
				orgID = id
			}
		}
	}
	var req requests.IngestLogRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	event, err := h.service.IngestLog(c.Context(), orgID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INGESTION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusCreated, "Event ingested", event.ToResponse())
}

// @Summary Bulk Ingest Logs
// @Description Bulk ingests raw security events for a data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param request body requests.BulkIngestLogsRequest true "Bulk Ingest Logs Request"
// @Success 200 {object} response.Response
// @Router /logs/bulk-ingest [post]
func (h *EventHandler) BulkIngestLogs(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	// Allow override for testing
	if localOrgID := c.Locals("orgID"); localOrgID != nil {
		if id, ok := localOrgID.(uuid.UUID); ok && id != uuid.Nil {
			orgID = id
		}
	}
	var req requests.BulkIngestLogsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	result, err := h.service.BulkIngestLogs(c.Context(), orgID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "BULK_INGESTION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Bulk ingestion completed", result)
}
