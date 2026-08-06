package domain

import (
	"database/sql"
	"time"
)

type UserSession struct {
	id               string
	userID           string
	organizationID   sql.NullString
	sessionTokenHash string
	ipAddress        sql.NullString
	userAgent        sql.NullString
	deviceInfo       sql.NullString
	location         sql.NullString
	isCurrent        bool
	lastActiveAt     time.Time
	expiresAt        time.Time
	createdAt        time.Time
	revokedAt        sql.NullTime
}

func NewUserSession(
	id string,
	userID string,
	organizationID sql.NullString,
	sessionTokenHash string,
	ipAddress sql.NullString,
	userAgent sql.NullString,
	deviceInfo sql.NullString,
	location sql.NullString,
	isCurrent bool,
	expiresAt time.Time,
	now time.Time,
) *UserSession {
	return &UserSession{
		id:               id,
		userID:           userID,
		organizationID:   organizationID,
		sessionTokenHash: sessionTokenHash,
		ipAddress:        ipAddress,
		userAgent:        userAgent,
		deviceInfo:       deviceInfo,
		location:         location,
		isCurrent:        isCurrent,
		lastActiveAt:     now,
		expiresAt:        expiresAt,
		createdAt:        now,
	}
}

func ReconstituteUserSession(
	id string,
	userID string,
	organizationID sql.NullString,
	sessionTokenHash string,
	ipAddress sql.NullString,
	userAgent sql.NullString,
	deviceInfo sql.NullString,
	location sql.NullString,
	isCurrent bool,
	lastActiveAt time.Time,
	expiresAt time.Time,
	createdAt time.Time,
	revokedAt sql.NullTime,
) *UserSession {
	return &UserSession{
		id:               id,
		userID:           userID,
		organizationID:   organizationID,
		sessionTokenHash: sessionTokenHash,
		ipAddress:        ipAddress,
		userAgent:        userAgent,
		deviceInfo:       deviceInfo,
		location:         location,
		isCurrent:        isCurrent,
		lastActiveAt:     lastActiveAt,
		expiresAt:        expiresAt,
		createdAt:        createdAt,
		revokedAt:        revokedAt,
	}
}

// Getters

func (s *UserSession) ID() string               { return s.id }
func (s *UserSession) UserID() string           { return s.userID }
func (s *UserSession) OrganizationID() sql.NullString { return s.organizationID }
func (s *UserSession) SessionTokenHash() string { return s.sessionTokenHash }
func (s *UserSession) IPAddress() sql.NullString { return s.ipAddress }
func (s *UserSession) UserAgent() sql.NullString { return s.userAgent }
func (s *UserSession) DeviceInfo() sql.NullString { return s.deviceInfo }
func (s *UserSession) Location() sql.NullString { return s.location }
func (s *UserSession) IsCurrent() bool          { return s.isCurrent }
func (s *UserSession) LastActiveAt() time.Time  { return s.lastActiveAt }
func (s *UserSession) ExpiresAt() time.Time     { return s.expiresAt }
func (s *UserSession) CreatedAt() time.Time     { return s.createdAt }
func (s *UserSession) RevokedAt() sql.NullTime  { return s.revokedAt }

// Business Rules

func (s *UserSession) Revoke() {
	s.revokedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

func (s *UserSession) Touch(now time.Time) {
	s.lastActiveAt = now
}
