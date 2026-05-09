package entra

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type EntraProvider struct {
	client *resty.Client
	TenantID     string
	ClientID     string
	ClientSecret string
}

func NewEntraProvider(tenantID, clientID, clientSecret string) *EntraProvider {
	client := resty.New()
	return &EntraProvider{
		client: client,
		TenantID: tenantID,
		ClientID: clientID,
		ClientSecret: clientSecret,
	}
}

func (e *EntraProvider) Verify(ctx context.Context) error {
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", e.TenantID)

	resp, err := e.client.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"client_id":     e.ClientID,
			"client_secret": e.ClientSecret,
			"grant_type":    "client_credentials",
			"scope":         "https://graph.microsoft.com/.default",
		}).
		Post(url)

	if err != nil {
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	return fmt.Errorf("failed to authenticate with entra: %s", resp.Status())
}