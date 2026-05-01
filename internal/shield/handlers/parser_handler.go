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

// @Summary Get Parser Summary
// @Description Returns parser library summary metrics including total parsers, active parsers, error rate, and last updated timestamp.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/summary [get]
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

// @Summary List Custom Parsers
// @Description Returns custom parsers with status, parsed event count, error rate, owner, and last updated time.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search parser name"
// @Param status query string false "Parser status"
// @Param data_source query string false "Data source"
// @Param parser_type query string false "Parser type"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers [get]
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

// @Summary Get Parser Detail
// @Description Returns full parser details including parser logic, tags, field mappings, status, and metrics.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Success 200 {object} response.Response
// @Routermiddl/wares.Gelogs-data/parsers/{id} [get]
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

// @Summary Create Custom Parser
// @Description Creates a custom parser for extracting, structuring, and normalizing raw logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param request body requests.CreateParserRequest true "Create Parser Request"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers [post]
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

// @Summary Update Custom Parser
// @Description Updates parser configuration, logic, mappings, tags, or status. Creates a parser version before saving changes.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Param request body requests.UpdateParserRequest true "Update Parser Request"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id} [patch]
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

// @Summary Test Parser
// @Description Tests a parser against a sample log and returns parsed output, normalized output, errors, and field mappings.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Param request body requests.TestParserRequest true "Test Parser Request"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id}/test [post]
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

// @Summary Preview Parser
// @Description Previews parser output before saving a new parser.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param request body requests.PreviewParserRequest true "Preview Parser Request"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/preview [post]
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

// @Summary Enable Parser
// @Description Enables a custom parser for future log ingestion.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id}/enable [post]
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

// @Summary Disable Parser
// @Description Disables a custom parser without deleting it.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id}/disable [post]
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

// @Summary Validate Parser
// @Description Queues validation for a parser against recent logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id}/validate [post]
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

// @Summary Validate All Parsers
// @Descrimiddlewares.Gtion Queues validation for all custom parsers.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/validate [post]
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

// @Summary Import Parser
// @Description Imports a custom parser definition from JSON payload or uploaded parser configuration.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param middlewares.Gequest body requests.ImportParserRequest true "Import Parser Request"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/import [post]
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

// @Summary Export Parser
// @Description Exports a parser definition as JSON.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param id path string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/{id}/export [get]
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

// @Summary List Sample Logs
// @Description Returns sample raw logs for parser testing and preview.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Param source_id query string false "Data source ID"
// @Param parser_id query string false "Parser ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response
// @Router /logs-data/parsers/sample-logs [get]
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
