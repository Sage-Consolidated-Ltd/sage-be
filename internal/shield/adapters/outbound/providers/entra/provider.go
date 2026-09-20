package entra

import (
	"sage-backend/internal/shield/domain"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type EntraProvider struct {
	RestyClient  *resty.Client
	TenantID     string
	ClientID     string
	ClientSecret string

	// Rate limiting
	BackoffSec     int
	MaxBackoffSec  int
	ConsecutiveErr int

	// Token management
	AccessToken    string
	TokenExpiresAt time.Time
	TokenMutex     sync.RWMutex

	BaseUrl    string
	Checkpoint *domain.Checkpoint
}

