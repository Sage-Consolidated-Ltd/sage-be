package entra

import (
	"context"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/go-resty/resty/v2"
)

func NewEntraProvider(
	tenantID, clientID, clientSecret string,
	client *resty.Client,
	checkpoint *domain.Checkpoint,
) *EntraProvider {
	if client == nil {
		client = resty.New()
	}
	client.SetTimeout(30 * time.Second)

	return &EntraProvider{
		RestyClient:    client,
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		BackoffSec:     30,
		MaxBackoffSec:  600,
		BaseUrl:        "https://graph.microsoft.com/v1.0",
		Checkpoint:     checkpoint,
	}
}

func (e *EntraProvider) Verify(ctx context.Context) error {
	_, err := e.getToken(ctx)
	return err
}
