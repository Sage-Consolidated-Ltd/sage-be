package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/redis/go-redis/v9"
)

// CorrelationStore implements ports.outbound.CorrelationStore backed by Redis Sorted Sets (ZSET).
type CorrelationStore struct {
	client redis.Cmdable
	prefix string
}

// NewCorrelationStore creates a Redis-backed CorrelationStore instance.
func NewCorrelationStore(client redis.Cmdable) outbound.CorrelationStore {
	return &CorrelationStore{
		client: client,
		prefix: "corr:",
	}
}

// AddEvent serializes the security event and adds it to a Redis sorted set scored by occurrence timestamp.
func (s *CorrelationStore) AddEvent(ctx context.Context, key string, event *domain.SecurityEvent, ttl time.Duration) error {
	if s.client == nil || event == nil || key == "" {
		return nil
	}

	redisKey := s.prefix + key
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for correlation store: %w", err)
	}

	eventTime := event.OccurredAt
	if eventTime.IsZero() {
		eventTime = time.Now()
	}
	score := float64(eventTime.UnixNano())

	// Pipeline ZADD, ZREMRANGEBYSCORE, and EXPIRE atomically
	pipe := s.client.TxPipeline()
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  score,
		Member: string(eventBytes),
	})

	cutoff := float64(time.Now().Add(-ttl).UnixNano())
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("(%f", cutoff))
	pipe.Expire(ctx, redisKey, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis transaction failed during AddEvent: %w", err)
	}

	return nil
}

// GetEvents retrieves all active events stored under key within the specified window duration.
func (s *CorrelationStore) GetEvents(ctx context.Context, key string, window time.Duration) ([]*domain.SecurityEvent, error) {
	if s.client == nil || key == "" {
		return []*domain.SecurityEvent{}, nil
	}

	redisKey := s.prefix + key
	now := time.Now()
	cutoffNano := now.Add(-window).UnixNano()
	minScore := strconv.FormatInt(cutoffNano, 10)

	// Clean up stale elements older than window first
	_ = s.client.ZRemRangeByScore(ctx, redisKey, "-inf", "("+minScore).Err()

	// Fetch elements in window range [minScore, +inf]
	rawEvents, err := s.client.ZRangeByScore(ctx, redisKey, &redis.ZRangeBy{
		Min: minScore,
		Max: "+inf",
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []*domain.SecurityEvent{}, nil
		}
		return nil, fmt.Errorf("failed to fetch correlation events from redis: %w", err)
	}

	events := make([]*domain.SecurityEvent, 0, len(rawEvents))
	for _, raw := range rawEvents {
		var evt domain.SecurityEvent
		if err := json.Unmarshal([]byte(raw), &evt); err == nil {
			events = append(events, &evt)
		}
	}

	return events, nil
}

// GetCount returns the number of events stored under key within the specified window duration.
func (s *CorrelationStore) GetCount(ctx context.Context, key string, window time.Duration) (int, error) {
	if s.client == nil || key == "" {
		return 0, nil
	}

	redisKey := s.prefix + key
	now := time.Now()
	cutoffNano := now.Add(-window).UnixNano()
	minScore := strconv.FormatInt(cutoffNano, 10)

	// Clean up stale elements older than window
	_ = s.client.ZRemRangeByScore(ctx, redisKey, "-inf", "("+minScore).Err()

	count, err := s.client.ZCount(ctx, redisKey, minScore, "+inf").Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count correlation events in redis: %w", err)
	}

	return int(count), nil
}

// Clear deletes the Redis key associated with the correlation key.
func (s *CorrelationStore) Clear(ctx context.Context, key string) error {
	if s.client == nil || key == "" {
		return nil
	}

	redisKey := s.prefix + key
	return s.client.Del(ctx, redisKey).Err()
}
