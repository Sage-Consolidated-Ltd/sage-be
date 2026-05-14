package okta

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type OktaProvider struct {
	client *resty.Client
	Domain string
	Token  string
}

func NewOktaProvider(domain, token string) *OktaProvider {
	client := resty.New().
		SetBaseURL(domain).
		SetAuthScheme("SSWS").
		SetAuthToken(token).
		SetHeader("Accept", "application/json")

	return &OktaProvider{
		client: client,
		Domain: domain,
		Token:  token,
	}
}

func (o *OktaProvider) Verify(ctx context.Context) error {
	resp, err := o.client.R().
		SetContext(ctx).
		SetQueryParam("limit", "1").
		Get("/api/v1/users")

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("invalid okta token")
	default:
		return fmt.Errorf("verification failed: %s", resp.Status())
	}
}
