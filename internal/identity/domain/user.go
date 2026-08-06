package domain

import (
	"database/sql"
	"time"
)

type User struct {
	id                string
	firstName         string
	lastName          string
	email             Email
	phoneNumber       sql.NullString
	backupEmail       sql.NullString
	passwordHash      PasswordHash
	passwordChangedAt sql.NullTime
	role              UserRole
	isVerified        bool
	twoFactorSecret   sql.NullString
	twoFactorEnabled  bool
	timeZone          sql.NullString
	avatarURL         sql.NullString
	lastLoginAt       sql.NullTime
	createdAt         time.Time
	updatedAt         sql.NullTime
	deletedAt         sql.NullTime
}

func NewUser(
	id string,
	firstName string,
	lastName string,
	email Email,
	passwordHash PasswordHash,
	role UserRole,
	now time.Time,
) *User {
	return &User{
		id:                id,
		firstName:         firstName,
		lastName:          lastName,
		email:             email,
		passwordHash:      passwordHash,
		passwordChangedAt: sql.NullTime{Time: now, Valid: true},
		role:              role,
		isVerified:        false,
		createdAt:         now,
	}
}

func ReconstituteUser(
	id string,
	firstName string,
	lastName string,
	email Email,
	phoneNumber sql.NullString,
	backupEmail sql.NullString,
	passwordHash PasswordHash,
	passwordChangedAt sql.NullTime,
	role UserRole,
	isVerified bool,
	twoFactorSecret sql.NullString,
	twoFactorEnabled bool,
	timeZone sql.NullString,
	avatarURL sql.NullString,
	lastLoginAt sql.NullTime,
	createdAt time.Time,
	updatedAt sql.NullTime,
	deletedAt sql.NullTime,
) *User {
	return &User{
		id:                id,
		firstName:         firstName,
		lastName:          lastName,
		email:             email,
		phoneNumber:       phoneNumber,
		backupEmail:       backupEmail,
		passwordHash:      passwordHash,
		passwordChangedAt: passwordChangedAt,
		role:              role,
		isVerified:        isVerified,
		twoFactorSecret:  twoFactorSecret,
		twoFactorEnabled: twoFactorEnabled,
		timeZone:         timeZone,
		avatarURL:        avatarURL,
		lastLoginAt:      lastLoginAt,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
	}
}

// Getters

func (u *User) ID() string                { return u.id }
func (u *User) FirstName() string         { return u.firstName }
func (u *User) LastName() string          { return u.lastName }
func (u *User) Email() Email              { return u.email }
func (u *User) PhoneNumber() sql.NullString { return u.phoneNumber }
func (u *User) BackupEmail() sql.NullString { return u.backupEmail }
func (u *User) PasswordHash() PasswordHash { return u.passwordHash }
func (u *User) PasswordChangedAt() sql.NullTime { return u.passwordChangedAt }
func (u *User) Role() UserRole           { return u.role }
func (u *User) IsVerified() bool         { return u.isVerified }
func (u *User) TwoFactorSecret() sql.NullString  { return u.twoFactorSecret }
func (u *User) TwoFactorEnabled() bool { return u.twoFactorEnabled }
func (u *User) TimeZone() sql.NullString { return u.timeZone }
func (u *User) AvatarURL() sql.NullString { return u.avatarURL }
func (u *User) LastLoginAt() sql.NullTime { return u.lastLoginAt }
func (u *User) CreatedAt() time.Time     { return u.createdAt }
func (u *User) UpdatedAt() sql.NullTime   { return u.updatedAt }
func (u *User) DeletedAt() sql.NullTime   { return u.deletedAt }

// Business Rules / State Mutations

func (u *User) Verify() {
	u.isVerified = true
	u.updatedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

func (u *User) ChangePassword(newHash PasswordHash) {
	u.passwordHash = newHash
	u.updatedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

func (u *User) EnableTwoFactor(secret string) {
	u.twoFactorEnabled = true
	u.twoFactorSecret = sql.NullString{String: secret, Valid: true}
	u.updatedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

func (u *User) DisableTwoFactor() {
	u.twoFactorEnabled = false
	u.twoFactorSecret = sql.NullString{Valid: false}
	u.updatedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

func (u *User) UpdateProfile(firstName, lastName string, timeZone sql.NullString) {
	if firstName != "" {
		u.firstName = firstName
	}
	if lastName != "" {
		u.lastName = lastName
	}
	if timeZone.Valid {
		u.timeZone = timeZone
	}
	u.updatedAt = sql.NullTime{Time: time.Now(), Valid: true}
}

type ExternalUser struct {
	ID        string
	Email     Email
	FirstName string
	LastName  string
	AvatarURL string
}
