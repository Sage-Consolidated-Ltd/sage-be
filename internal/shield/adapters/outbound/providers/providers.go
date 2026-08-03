package providers

import (
	"context"
	"fmt"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/adapters/outbound/providers/entra"
	"strings"

	"sage-backend/internal/shield/adapters/outbound/providers/okta"

	"github.com/go-resty/resty/v2"
)

type Provider interface {
	Verify(ctx context.Context) error
	Collect(ctx context.Context, limit int) ([]domain.NormalizedEvent, error)
}

type OktaCredentials struct {
	Domain     string             `json:"domain"`
	Token      string             `json:"token"`
	Checkpoint *domain.Checkpoint `json:"checkpoint,omitempty"`
}

type EntraCredentials struct {
	TenantID     string             `json:"tenant_id"`
	ClientID     string             `json:"client_id"`
	ClientSecret string             `json:"client_secret"`
	Checkpoint   *domain.Checkpoint `json:"checkpoint,omitempty"`
}

func NewProvider(
	provider string,
	providerCreds any,
	client *resty.Client,
) (Provider, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))

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

		return okta.NewOktaProvider(creds.Domain, creds.Token, creds.Checkpoint), nil

	case "entra":
		creds, ok := providerCreds.(EntraCredentials)
		if !ok {
			return nil, fmt.Errorf("invalid entra config")
		}

		if creds.TenantID == "" || creds.ClientID == "" || creds.ClientSecret == "" {
			return nil, fmt.Errorf("missing entra config")
		}

		entraProvider, err := entra.NewEntraProvider(
			creds.TenantID,
			creds.ClientID,
			creds.ClientSecret,
			client,
			"redis://localhost:6379/0",
			300,
			creds.Checkpoint,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create entra provider: %w", err)
		}

		return entraProvider, nil

	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}

func LaunchProviderSync(provider string, credentials map[string]string, checkpoint *domain.Checkpoint, client *resty.Client) (Provider, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))

	switch provider {

	case "okta":
		domain := credentials["domain"]
		token := credentials["token"]

		if domain == "" || token == "" {
			return nil, fmt.Errorf(
				"missing okta config",
			)
		}

		// normalize domain
		if !strings.HasPrefix(domain, "http") {
			domain = "https://" + domain
		}

		domain = strings.TrimRight(domain, "/")

		return NewProvider(provider,
			OktaCredentials{
				Domain:     domain,
				Token:      token,
				Checkpoint: checkpoint,
			},
			client,
		)

	case "entra":
		tenantID := credentials["tenant_id"]
		clientID := credentials["client_id"]
		clientSecret := credentials["client_secret"]

		if tenantID == "" ||
			clientID == "" ||
			clientSecret == "" {
			return nil, fmt.Errorf(
				"missing entra config",
			)
		}

		return NewProvider(
			provider,
			EntraCredentials{
				TenantID:     credentials["tenant_id"],
				ClientID:     credentials["client_id"],
				ClientSecret: credentials["client_secret"],
				Checkpoint:   checkpoint,
			},
			client,
		)

	default:
		return nil, fmt.Errorf(
			"unsupported provider",
		)
	}
}
