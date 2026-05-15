package entra

import (
	"sage-backend/internal/shield/models"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
)

type EntraProvider struct {
	RestyClient     *resty.Client
	TenantID        string
	ClientID        string
	ClientSecret    string
	RedisClient     *redis.Client
	PollIntervalSec int

	// Rate limiting
	BackoffSec     int
	MaxBackoffSec  int
	ConsecutiveErr int

	// Token management
	AccessToken    string
	TokenExpiresAt time.Time
	TokenMutex     sync.RWMutex

	QueueKey      string
	CheckpointKey string
	DlqKey        string

	BaseUrl    string
	Checkpoint *models.Checkpoint
}
