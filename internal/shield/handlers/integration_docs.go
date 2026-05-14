package handlers

import (
	_ "sage-backend/internal/shared/response"
	_ "sage-backend/internal/shield/requests"
)

// @Summary Create Integration
// @Description Creates a new integration with the specified provider and configuration.
// @Tags Integrations
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body requests.CreateIntegrationRequest true "Integration creation payload"
// @Success 201 {object} response.Response
// @Router /integrations [post]
func _IntegrateDataSource() {}
