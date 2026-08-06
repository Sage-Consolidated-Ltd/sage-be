package integration

import (
	"context"
	"testing"

	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db/redis"

	redisClient "github.com/redis/go-redis/v9"
)

func testRedis(t *testing.T) *redisClient.Client {
	t.Helper()

	cfg := config.SetupTestConfig()

	rdb, err := redis.LaunchRedis(&cfg.BaseConfig)
	if err != nil {
		t.Fatalf("failed to connect to test redis: %v", err)
	}

	cleanupRedis(t, rdb)

	t.Cleanup(func() {
		if err := rdb.Close(); err != nil {
			t.Errorf("failed to close redis: %v", err)
		}
	})

	return rdb
}

func cleanupRedis(t *testing.T, rdb *redisClient.Client) {
	t.Helper()

	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush redis DB: %v", err)
	}
}
