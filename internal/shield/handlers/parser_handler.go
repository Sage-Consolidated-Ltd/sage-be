package handlers

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"
	"sage-backend/internal/shield/requests"
	"sage-backend/internal/shield/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ParserHandler struct {
	service services.ParserServiceInt
}

func NewParserHandler(service services.ParserServiceInt) *ParserHandler {
	return &ParserHandler{service: service}
}

func (h *ParserHandler) GetParserSummary(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	total, active, errorRate, lastUpdated, err := h.service.GetParserSummary(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_SUMMARY", err.Error())
	}
	summary := map[string]interface{}{
		"total_custom_parsers": total,
		"active_parsers":       active,
		"error_rate_24h":       errorRate,
		"last_updated_at":      lastUpdated,
	}
	return response.JSON(c, fiber.StatusOK, "Parser summary", summary)
}
func (h *ParserHandler) ListParsers(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	filters := map[string]interface{}{
		"status":      c.Query("status"),
		"parser_type": c.Query("parser_type"),
		"search":      c.Query("search"),
	}
	parsers, total, err := h.service.ListParsers(c.Context(), orgID, filters, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_LIST_PARSERS", err.Error())
	}
	items := make([]interface{}, len(parsers))
	for i, p := range parsers {
		// resolve owner name? skip for now.
		items[i] = p.ToResponse(nil)
	}
	paginated := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Parsers retrieved", paginated)
}
func (h *ParserHandler) GetParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	parser, err := h.service.GetParser(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "PARSER_NOT_FOUND", err.Error())
	}
	// Fetch owner name if needed; omitted
	return response.JSON(c, fiber.StatusOK, "Parser details", parser.ToResponse(nil))
}
func (h *ParserHandler) CreateParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	var req requests.CreateParserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	parser := &models.Parser{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		SourceID:       req.DataSourceID,
		ParserType:     req.ParserType,
		Status:         types.ParserStatusActive,
		Tags:           req.Tags,
		Logic:          req.Logic,
		Mappings:       req.Mappings,
	}
	if err := h.service.CreateParser(c.Context(), parser); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusCreated, "Parser created", parser.ToResponse(nil))
}
func (h *ParserHandler) UpdateParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	var req requests.UpdateParserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	// Fetch existing parser
	existing, err := h.service.GetParser(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "PARSER_NOT_FOUND", err.Error())
	}
	// Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.DataSourceID != nil {
		existing.SourceID = req.DataSourceID
	}
	if req.ParserType != nil {
		existing.ParserType = *req.ParserType
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Logic != nil {
		existing.Logic = req.Logic
	}
	if req.Mappings != nil {
		existing.Mappings = req.Mappings
	}
	// Get actor user ID from context for change log
	actorUserID := middlewares.GetUserID(c) // implement GetUserID in middleware (maybe exists)
	changeNote := "Updated parser"
	if err := h.service.UpdateParser(c.Context(), existing, changeNote, actorUserID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Parser updated", existing.ToResponse(nil))
}
func (h *ParserHandler) TestParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	var req requests.TestParserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	result, err := h.service.TestParser(c.Context(), id, orgID, req.SampleLog, req.RawPayload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "TEST_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Test result", result)
}
func (h *ParserHandler) PreviewParser(c *fiber.Ctx) error {
	var req requests.PreviewParserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	result, err := h.service.PreviewParser(c.Context(), req.ParserType, req.Logic, req.Mappings, req.SampleLog, req.RawPayload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "PREVIEW_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Preview result", result)
}
func (h *ParserHandler) EnableParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	if err := h.service.EnableParser(c.Context(), id, orgID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "ENABLE_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Parser enabled", nil)
}
func (h *ParserHandler) DisableParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	if err := h.service.DisableParser(c.Context(), id, orgID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DISABLE_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Parser disabled", nil)
}
func (h *ParserHandler) ValidateParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	result, err := h.service.ValidateParser(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "VALIDATION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Validation queued", result)
}
func (h *ParserHandler) ValidateAllParsers(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	result, err := h.service.ValidateAllParsers(c.Context(), orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "VALIDATION_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Validation queued for all parsers", result)
}
func (h *ParserHandler) ImportParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	var req requests.ImportParserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	parser := &models.Parser{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		SourceID:       req.DataSourceID,
		ParserType:     req.ParserType,
		Status:         req.Status,
		Tags:           req.Tags,
		Logic:          req.Logic,
		Mappings:       req.Mappings,
	}
	if err := h.service.ImportParser(c.Context(), parser); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "IMPORT_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusCreated, "Parser imported", parser.ToResponse(nil))
}
func (h *ParserHandler) ExportParser(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", err.Error())
	}
	parser, err := h.service.ExportParser(c.Context(), id, orgID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "EXPORT_FAILED", err.Error())
	}
	return response.JSON(c, fiber.StatusOK, "Parser exported", parser)
}
func (h *ParserHandler) ListSampleLogs(c *fiber.Ctx) error {
	orgID := middlewares.GetOrgID(c)
	if orgID == uuid.Nil {
		orgID = c.Locals("orgID").(uuid.UUID)
	}
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 25)
	var srcID, prsrID *uuid.UUID
	if s := c.Query("source_id"); s != "" {
		if uid, err := uuid.Parse(s); err == nil {
			srcID = &uid
		}
	}
	if s := c.Query("parser_id"); s != "" {
		if uid, err := uuid.Parse(s); err == nil {
			prsrID = &uid
		}
	}
	events, total, err := h.service.ListSampleLogs(c.Context(), srcID, prsrID, orgID, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_GET_SAMPLE_LOGS", err.Error())
	}
	items := make([]interface{}, len(events))
	for i, e := range events {
		items[i] = e.ToResponse()
	}
	paginated := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	return response.JSON(c, fiber.StatusOK, "Sample logs", paginated)
}
