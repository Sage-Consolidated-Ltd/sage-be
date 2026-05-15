package okta

import (
	"sage-backend/internal/shield/models"

	"github.com/go-resty/resty/v2"
)

type OktaProvider struct {
	Domain      string
	ApiToken    string
	RestyClient *resty.Client
	Checkpoint  *models.Checkpoint
}
