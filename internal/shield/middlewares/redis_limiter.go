<<<<<<< HEAD
package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiterStore struct {
	client *redis.Client
}

func NewRedisLimiterStore(client *redis.Client) *RedisLimiterStore {
	return &RedisLimiterStore{
		client: client,
	}
}

func (r *RedisLimiterStore) Increment(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (int, time.Time, error) {

	now := time.Now().UTC()
	windowStart := now.Truncate(window)

	redisKey := fmt.Sprintf(
		"rate_limit:%s:%d",
		key,
		windowStart.Unix(),
	)

	pipe := r.client.TxPipeline()

	countCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, time.Time{}, err
	}

	reset := windowStart.Add(window)

	return int(countCmd.Val()), reset, nil
=======
package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiterStore struct {
	client *redis.Client
}

func NewRedisLimiterStore(client *redis.Client) *RedisLimiterStore {
	return &RedisLimiterStore{
		client: client,
	}
}

func (r *RedisLimiterStore) Increment(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (int, time.Time, error) {

	now := time.Now().UTC()
	windowStart := now.Truncate(window)

	redisKey := fmt.Sprintf(
		"rate_limit:%s:%d",
		key,
		windowStart.Unix(),
	)

	pipe := r.client.TxPipeline()

	countCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, time.Time{}, err
	}

	reset := windowStart.Add(window)

	return int(countCmd.Val()), reset, nil
>>>>>>> 3cbfcfda147bdc68807f3505d9e623e7af4a13bc
}