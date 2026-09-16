package providers

import (
	"testing"

	"sage-backend/internal/shield/domain"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider_Entra(t *testing.T) {
	client := resty.New()

	t.Run("successfully creates Entra provider without requiring Redis", func(t *testing.T) {
		creds := EntraCredentials{
			TenantID:     "tenant-123",
			ClientID:     "client-456",
			ClientSecret: "secret-789",
		}

		provider, err := NewProvider("entra", creds, client)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("returns error when entra credentials are missing required fields", func(t *testing.T) {
		creds := EntraCredentials{
			TenantID: "",
			ClientID: "client-456",
		}

		provider, err := NewProvider("entra", creds, client)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "missing entra config")
	})

	t.Run("returns error on invalid entra credential type", func(t *testing.T) {
		provider, err := NewProvider("entra", "invalid-type", client)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "invalid entra config")
	})
}

func TestNewProvider_Okta(t *testing.T) {
	client := resty.New()

	t.Run("successfully creates Okta provider", func(t *testing.T) {
		creds := OktaCredentials{
			Domain: "example.okta.com",
			Token:  "test-token",
		}

		provider, err := NewProvider("okta", creds, client)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("returns error when okta credentials are missing required fields", func(t *testing.T) {
		creds := OktaCredentials{
			Domain: "",
			Token:  "",
		}

		provider, err := NewProvider("okta", creds, client)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "missing okta config")
	})
}

func TestLaunchProviderSync(t *testing.T) {
	client := resty.New()
	checkpoint := &domain.Checkpoint{}

	t.Run("successfully launches Entra provider sync", func(t *testing.T) {
		creds := map[string]string{
			"tenant_id":     "tenant-123",
			"client_id":     "client-456",
			"client_secret": "secret-789",
		}

		provider, err := LaunchProviderSync("entra", creds, checkpoint, client)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("returns error when Entra credentials missing in LaunchProviderSync", func(t *testing.T) {
		creds := map[string]string{
			"tenant_id": "tenant-123",
		}

		provider, err := LaunchProviderSync("entra", creds, checkpoint, client)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "missing entra config")
	})

	t.Run("returns error for unsupported provider", func(t *testing.T) {
		creds := map[string]string{}
		provider, err := LaunchProviderSync("unsupported", creds, checkpoint, client)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported provider")
	})
}
