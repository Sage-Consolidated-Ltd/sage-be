package postgres

import (
	"database/sql"
	"sage-backend/internal/identity/domain"
	orgDomain "sage-backend/internal/organization/domain"
	"time"
)

type userModel struct {
	ID                string         `db:"id"`
	FirstName         string         `db:"first_name"`
	LastName          string         `db:"last_name"`
	Email             string         `db:"email"`
	PhoneNumber       sql.NullString `db:"phone_number"`
	BackupEmail       sql.NullString `db:"backup_email"`
	IsVerified        bool           `db:"is_verified"`
	PasswordHash      string         `db:"password_hash"`
	PasswordChangedAt sql.NullTime   `db:"password_changed_at"`
	Role              string         `db:"role"`
	TwoFactorSecret   sql.NullString `db:"two_factor_secret"`
	TwoFactorEnabled  bool           `db:"two_factor_enabled"`
	TimeZone          sql.NullString `db:"time_zone"`
	AvatarURL         sql.NullString `db:"avatar_url"`
	LastLoginAt       sql.NullTime   `db:"last_login_at"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         sql.NullTime   `db:"updated_at"`
	DeletedAt         sql.NullTime   `db:"deleted_at"`
}

func (m userModel) ToDomain() (*domain.User, error) {
	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}
	role, err := domain.NewUserRole(m.Role)
	if err != nil {
		role = domain.MustNewUserRole("user")
	}
	hash := domain.NewPasswordHash(m.PasswordHash)
	return domain.ReconstituteUser(
		m.ID,
		m.FirstName,
		m.LastName,
		email,
		m.PhoneNumber,
		m.BackupEmail,
		hash,
		m.PasswordChangedAt,
		role,
		m.IsVerified,
		m.TwoFactorSecret,
		m.TwoFactorEnabled,
		m.TimeZone,
		m.AvatarURL,
		m.LastLoginAt,
		m.CreatedAt,
		m.UpdatedAt,
		m.DeletedAt,
	), nil
}

type organizationModel struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	Slug                 string         `db:"slug"`
	OwnerID              string         `db:"owner_id"`
	Industry             string         `db:"industry"`
	Country              sql.NullString `db:"country"`
	Timezone             string         `db:"timezone"`
	RiskThresholdDefault int            `db:"risk_threshold_default"`
	Role                 string         `db:"role"`
	Status               string         `db:"status"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	DeletedAt            sql.NullTime   `db:"deleted_at"`
}

func (m organizationModel) ToDomain() orgDomain.Organization {
	return orgDomain.Organization{
		ID:                   m.ID,
		Name:                 m.Name,
		Slug:                 m.Slug,
		OwnerID:              m.OwnerID,
		Industry:             m.Industry,
		Country:              m.Country,
		Timezone:             m.Timezone,
		RiskThresholdDefault: m.RiskThresholdDefault,
		Role:                 m.Role,
		Status:               m.Status,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		DeletedAt:            m.DeletedAt,
	}
}
