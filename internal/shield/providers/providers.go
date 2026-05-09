package providers

import (
	"context"
	"fmt"
	"sage-backend/internal/shield/providers/entra"
	"sage-backend/internal/shield/providers/okta"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Provider interface {
	Verify(ctx context.Context) error
}

type OktaCredentials struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

type EntraCredentials struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewProvider(
	provider string, 
	providerCreds any, 
	client *resty.Client,
) (Provider, error) {

	switch provider {

	case "okta":
		creds, ok := providerCreds.(OktaCredentials) 
		if !ok {
			return nil, fmt.Errorf("invalid okta config")
		}

		if creds.Domain == "" || creds.Token == "" {
			return nil, fmt.Errorf("missing okta config")
		}

		// normalize domain
		if !strings.HasPrefix(creds.Domain, "http") {
			creds.Domain = "https://" + creds.Domain
		}
		creds.Domain = strings.TrimRight(creds.Domain, "/")

		return okta.NewOktaProvider(creds.Domain, creds.Token), nil

	case "entra":
		creds, ok := providerCreds.(EntraCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid entra config")
		}

		if creds.TenantID == "" || creds.ClientID == "" || creds.ClientSecret == "" {
			return nil, fmt.Errorf("missing entra config")
		}

		return entra.NewEntraProvider(
			creds.TenantID,
			creds.ClientID,
			creds.ClientSecret,
		), nil

	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}