package models

import (
	"database/sql"
	"time"
)

type UserSession struct {
	ID             string         `json:"id" db:"id"`
	UserID         string         `json:"user_id" db:"user_id"`
	OrganizationID sql.NullString `json:"organization_id" db:"organization_id"`
	SessionTokenHash string       `json:"-" db:"session_token_hash"`
	IPAddress      sql.NullString `json:"ip_address" db:"ip_address"`
	UserAgent      sql.NullString `json:"user_agent" db:"user_agent"`
	Location       sql.NullString `json:"location" db:"location"`
	IsRevoked      bool           `json:"is_revoked" db:"is_revoked"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	LastActiveAt   time.Time      `json:"last_active_at" db:"last_active_at"`
	ExpiresAt      time.Time      `json:"expires_at" db:"expires_at"`
}

type UserSessionResponse struct {
	SessionID    string    `json:"session_id"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Location     string    `json:"location,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsCurrent    bool      `json:"is_current"`
}

func (s *UserSession) ToResponse(currentSessionID string) *UserSessionResponse {
	resp := &UserSessionResponse{
		SessionID:    s.ID,
		CreatedAt:    s.CreatedAt,
		LastActiveAt: s.LastActiveAt,
		IsCurrent:    s.ID == currentSessionID,
	}
	if s.IPAddress.Valid {
		resp.IPAddress = s.IPAddress.String
	}
	if s.UserAgent.Valid {
		resp.UserAgent = s.UserAgent.String
	}
	if s.Location.Valid {
		resp.Location = s.Location.String
	}
	return resp
}
