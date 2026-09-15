package domain

import (
	"time"
)

type Admin struct {
	id             string
	email          Email
	passwordHash   string
	role           Role
	status         Status
	failedAttempts int
	lockedUntil    *time.Time
	lastLoginAt    *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func NewAdmin(
	id string,
	email Email,
	passwordHash string,
	role Role,
	now time.Time,
) *Admin {
	return &Admin{
		id:             id,
		email:          email,
		passwordHash:   passwordHash,
		role:           role,
		status:         StatusActive,
		failedAttempts: 0,
		lockedUntil:    nil,
		lastLoginAt:    nil,
		createdAt:      now,
		updatedAt:      now,
	}
}

func ReconstituteAdmin(
	id string,
	email Email,
	passwordHash string,
	role Role,
	status Status,
	failedAttempts int,
	lockedUntil *time.Time,
	lastLoginAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Admin {
	return &Admin{
		id:             id,
		email:          email,
		passwordHash:   passwordHash,
		role:           role,
		status:         status,
		failedAttempts: failedAttempts,
		lockedUntil:    lockedUntil,
		lastLoginAt:    lastLoginAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (a *Admin) ID() string {
	return a.id
}

func (a *Admin) Email() Email {
	return a.email
}

func (a *Admin) PasswordHash() string {
	return a.passwordHash
}

func (a *Admin) Role() Role {
	return a.role
}

func (a *Admin) Status() Status {
	return a.status
}

func (a *Admin) FailedAttempts() int {
	return a.failedAttempts
}

func (a *Admin) LockedUntil() *time.Time {
	return a.lockedUntil
}

func (a *Admin) LastLoginAt() *time.Time {
	return a.lastLoginAt
}

func (a *Admin) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Admin) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *Admin) CanLogin(now time.Time) error {
	if a.status == StatusSuspended {
		return ErrAccountSuspended
	}
	if a.status == StatusPermanentlyLocked {
		return ErrAccountPermanentlyLocked
	}
	if a.status == StatusTemporarilyLocked {
		if a.lockedUntil != nil && now.Before(*a.lockedUntil) {
			return ErrAccountTemporarilyLocked
		}
	}
	return nil
}

// RecordFailedAttempt tracks consecutive failed logins and updates the lockout state machine:
// - 5 attempts -> temporarily locked for 30 minutes
// - 10 attempts -> permanently locked
func (a *Admin) RecordFailedAttempt(now time.Time) (bool, Status) {
	a.failedAttempts++
	a.updatedAt = now

	if a.failedAttempts >= 10 {
		a.status = StatusPermanentlyLocked
		a.lockedUntil = nil
		return true, StatusPermanentlyLocked
	}

	if a.failedAttempts >= 5 {
		a.status = StatusTemporarilyLocked
		lockTime := now.Add(30 * time.Minute)
		a.lockedUntil = &lockTime
		return true, StatusTemporarilyLocked
	}

	return false, a.status
}

// RecordSuccessfulLogin clears failure counters and updates the last login timestamp.
func (a *Admin) RecordSuccessfulLogin(now time.Time) {
	a.failedAttempts = 0
	a.lockedUntil = nil
	if a.status == StatusTemporarilyLocked {
		a.status = StatusActive
	}
	a.lastLoginAt = &now
	a.updatedAt = now
}

// Unlock manually clears locks and restores the admin to active status (super admin action).
func (a *Admin) Unlock(now time.Time) {
	a.failedAttempts = 0
	a.lockedUntil = nil
	a.status = StatusActive
	a.updatedAt = now
}

// ChangePassword updates the password hash and resets lockout counters.
func (a *Admin) ChangePassword(newHash string, now time.Time) {
	a.passwordHash = newHash
	a.failedAttempts = 0
	a.lockedUntil = nil
	a.updatedAt = now
}
