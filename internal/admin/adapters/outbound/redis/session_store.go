package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/outbound"

	"github.com/redis/go-redis/v9"
)

type SessionPayload struct {
	SessionID string    `json:"session_id"`
	AdminID   string    `json:"admin_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	BoundIP   string    `json:"bound_ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client *redis.Client) outbound.SessionStore {
	return &RedisSessionStore{client: client}
}

var _ outbound.SessionStore = (*RedisSessionStore)(nil)

func sessionKey(sessionID string) string {
	return fmt.Sprintf("admin:session:%s", sessionID)
}

func (s *RedisSessionStore) CreateSession(ctx context.Context, session *outbound.AdminSession, ttl time.Duration) error {
	payload := SessionPayload{
		SessionID: session.SessionID,
		AdminID:   session.AdminID,
		Role:      session.Role,
		Email:     session.Email,
		BoundIP:   session.BoundIP,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, sessionKey(session.SessionID), raw, ttl).Err()
}

func (s *RedisSessionStore) GetSession(ctx context.Context, sessionID string) (*outbound.AdminSession, error) {
	raw, err := s.client.Get(ctx, sessionKey(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrSessionExpired
		}
		return nil, err
	}

	var payload SessionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return &outbound.AdminSession{
		SessionID: payload.SessionID,
		AdminID:   payload.AdminID,
		Role:      payload.Role,
		Email:     payload.Email,
		BoundIP:   payload.BoundIP,
		CreatedAt: payload.CreatedAt,
		ExpiresAt: payload.ExpiresAt,
	}, nil
}

func (s *RedisSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, sessionKey(sessionID)).Err()
}
