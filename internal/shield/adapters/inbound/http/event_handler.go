package http

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/ports/dto"
	"sage-backend/internal/shield/ports/inbound"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EventHandler struct {
	service inbound.LogsUseCase
}

func NewEventHandler(service inbound.LogsUseCase) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) SearchLogs(c *fiber.Ctx) error {
	var orgID uuid.UUID
	if localOrgID := c.Locals("orgID"); localOrgID != nil {
		if id, ok := localOrgID.(uuid.UUID); ok && id != uuid.Nil {
			orgID = id
		}
	}
	if orgID == uuid.Nil {
		var err error
		orgID, err = middlewares.GetOrgID(c)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		}
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)

	if q := c.Query("q"); q != "" {
		res, err := h.service.SearchLogsAST(c.Context(), orgID, q, pageSize)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_SEARCH_LOGS", err.Error())
		}
		return response.JSON(c, fiber.StatusOK, "Logs retrieved", res)
	}

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
		items[i] = dto.SecurityEventToResponse(log)
	}
	paginated := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Logs retrieved", paginated)
}
func (h *EventHandler) GetLogDetail(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	event, err := h.service.GetLogByID(c.Context(), orgID, id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "EVENT_NOT_FOUND", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Event retrieved", dto.SecurityEventToResponse(event))
}
func (h *EventHandler) IngestLog(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	var req dto.IngestLogRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	event, err := h.service.IngestLog(c.Context(), orgID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INGESTION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusCreated, "Event ingested", dto.SecurityEventToResponse(event))
}
func (h *EventHandler) BulkIngestLogs(c *fiber.Ctx) error {
	orgID, err := getOrgID(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	}
	var req dto.BulkIngestLogsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	result, err := h.service.BulkIngestLogs(c.Context(), orgID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "BULK_INGESTION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Bulk ingestion completed", result)
}
