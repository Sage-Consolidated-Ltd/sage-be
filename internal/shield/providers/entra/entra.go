package entra

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
)

func NewEntraProvider(tenantID, clientID, clientSecret string, client *resty.Client, redisURL string,
	pollIntervalSec int) (*EntraProvider, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}
	
	redisClient := redis.NewClient(opts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	client.SetTimeout(30 *time.Second)

	return &EntraProvider{
		RestyClient: client,
		TenantID: tenantID,
		ClientID: clientID,
		ClientSecret: clientSecret,
		RedisClient:     redisClient,
		PollIntervalSec: pollIntervalSec,
		BackoffSec:    30,
		MaxBackoffSec: 600,
		QueueKey:      "azure_ad:events",
		CheckpointKey: "azure_ad:checkpoint",
		DlqKey:        "azure_ad:dlq",
		BaseUrl: "https://graph.microsoft.com/v1.0",
	}, nil
}

func (e *EntraProvider) Verify(ctx context.Context) error {
	_, err := e.getToken(ctx)
	return err
}