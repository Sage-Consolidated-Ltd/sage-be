package outbound

import (
	"context"
	"time"
)

type AdminSession struct {
	SessionID string
	AdminID   string
	Role      string
	Email     string
	BoundIP   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore interface {
	CreateSession(ctx context.Context, session *AdminSession, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (*AdminSession, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
