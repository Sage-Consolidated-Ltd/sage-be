package requests

type CreateIntegrationRequest struct {
	Name           string `json:"name" validate:"required"`
	Provider       string `json:"provider" validate:"required"`
	ConnectionType string `json:"connection_type" validate:"required"`

	Okta  *OktaParams  `json:"okta,omitempty"`
	Entra *EntraParams `json:"entra,omitempty"`
}

type OktaParams struct {
	Domain string `json:"domain" validate:"required,url"`
	Token  string `json:"token" validate:"required"`
}

type EntraParams struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

type CredentialInput struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}
