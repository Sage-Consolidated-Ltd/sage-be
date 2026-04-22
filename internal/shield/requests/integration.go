package requests

import (
	"encoding/json"
	"fmt"
	"sage-backend/internal/shield/models"
)

type CreateIntegrationRequest struct {
	Name           string                 `json:"name" validate:"required"`
	Provider       string                 `json:"provider" validate:"required"`
	ConnectionType string                 `json:"connection_type" validate:"required,oneof=webhook kafka pubsub agent"`
	Config         map[string]interface{} `json:"config" validate:"required"`
	Credentials    []CredentialInput      `json:"credentials"`
}

func MapCreateIntegration(req CreateIntegrationRequest, tenantID string) models.Integration {
	configBytes, _ := json.Marshal(req.Config)

	return models.Integration{
		TenantId:       tenantID,
		Name:           req.Name,
		Provider:       req.Provider,
		ConnectionType: req.ConnectionType,
		Status:         "inactive",
		Config:         configBytes,
	}
}

func MapIntegrationResponse(payload *models.Integration) *models.IntegrationResponse {
    config := make(map[string]interface{})
    
    if len(payload.Config) > 0 {
        err := json.Unmarshal(payload.Config, &config)
        if err != nil {
            fmt.Printf("failed to unmarshal config: %v\n", err)
        }
    }

    return &models.IntegrationResponse{
        ID:             payload.ID,
        Name:           payload.Name,
        Provider:       payload.Provider,
        ConnectionType: payload.ConnectionType,
        Status:         payload.Status,
        Config:         config,
        CreatedAt:      payload.CreatedAt,
    }
}

type CredentialInput struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}
