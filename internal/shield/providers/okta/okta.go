package okta

import (
	"context"
	"fmt"
	"sage-backend/internal/shield/models"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func NewOktaProvider(domain, token string, checkpoint *models.Checkpoint) *OktaProvider {

	return &OktaProvider{
		RestyClient: resty.New().
			SetTimeout(30 * time.Second).
			SetBaseURL(domain),
		Domain:     domain,
		ApiToken:   token,
		Checkpoint: checkpoint,
	}
}

func (o *OktaProvider) Verify(ctx context.Context) error {
	resp, err := o.RestyClient.R().
		SetContext(ctx).
		SetQueryParam("limit", "1").
		SetHeader("Authorization", "SSWS "+strings.TrimSpace(o.ApiToken)).
		SetHeader("Accept", "application/json").
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
