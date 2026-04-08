package redis

import (
	"context"
	"fmt"
	"sage-backend/internal/shared/config"
	"time"

	// "github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// func GetAsynqOptions(cfg *config.Config) (asynq.RedisClientOpt, error) {
// 	opt, err := redis.ParseURL(cfg.RedisDbUrl)
// 	if err != nil {
// 		return asynq.RedisClientOpt{}, err
// 	}

// 	return asynq.RedisClientOpt{
// 		Addr:     opt.Addr,
// 		Password: opt.Password,
// 		DB:       opt.DB,
// 	}, nil
// }

func LaunchRedis(cfg *config.BaseConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisDbUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}
