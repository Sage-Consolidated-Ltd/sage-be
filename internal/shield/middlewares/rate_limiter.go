package middlewares

import (
	"context"
	"sage-backend/internal/shared/logger"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type bucket struct {
	count    int
	windowAt time.Time
}

type LimiterStore interface {
	Increment(
		ctx context.Context,
		key string,
		limit int,
		window time.Duration,
	) (count int, reset time.Time, err error)
}

type RateLimiter struct {
	rpm    int
	window time.Duration

	mu      sync.Mutex
	buckets map[string]*bucket

	store LimiterStore
	log   *logger.Logger
}

func NewRateLimiter(rpm int, log *logger.Logger, store LimiterStore) *RateLimiter {
	return &RateLimiter{
		rpm:     rpm,
		buckets: make(map[string]*bucket),
		store:   store,
		log:     log,
	}
}

func (rl *RateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var key string

		// authenticated user
		userID, ok := c.Locals("userID").(string)
		if ok && userID != "" {
			key = "user:" + userID + ":presign"
		} else {
			// fallback for unauthenticated routes
			key = "ip:" + realIP(c)
		}

		var (
			count int
			reset time.Time
		)

		if rl.store != nil {
			var err error

			count, reset, err = rl.store.Increment(
				c.Context(),
				key,
				rl.rpm,
				time.Minute,
			)
			if err != nil {
				rl.log.Error("rate limiter store error",
					zap.Error(err),
				)

				// fail open
				return c.Next()
			}
		} else {
			count = rl.inMemoryIncrement(key)
			reset = time.Now().Add(time.Minute)
		}

		remaining := max(0, rl.rpm-count)

		c.Set("X-RateLimit-Limit", itoa(rl.rpm))
		c.Set("X-RateLimit-Remaining", itoa(remaining))
		c.Set(
			"X-RateLimit-Reset",
			itoa(int(reset.Unix())),
		)

		if count >= rl.rpm {
			retryAfter := max(
				0,
				int(time.Until(reset).Seconds()),
			)

			c.Set(
				"Retry-After",
				itoa(retryAfter),
			)

			return c.Status(
				fiber.StatusTooManyRequests,
			).JSON(fiber.Map{
				"code":    "RATE_LIMITED",
				"message": "too many requests",
			})
		}

		return c.Next()
	}
}

func (rl *RateLimiter) inMemoryIncrement(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]

	if !ok || time.Since(b.windowAt) > rl.window {
		rl.buckets[key] = &bucket{
			count:    1,
			windowAt: time.Now(),
		}
		return 1
	}

	b.count++
	return b.count
}
