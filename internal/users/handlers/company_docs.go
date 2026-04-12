package handlers

import "sage-backend/internal/users/models"

type IndustriesResponse struct {
	Success bool                           `json:"success"`
	Message string                         `json:"message"`
	Data    []models.GetIndustriesResponse `json:"data"`
}

// @Summary Get Industries
// @Description Retrieve a list of industries for company profiles.
// @Tags Company
// @Accept json
// @Produce json
// @Success 200 {object} IndustriesResponse
// @Router /company/industries [get]
func _GetIndustries(){}