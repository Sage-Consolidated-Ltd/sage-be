package postgres

import (
	"time"
	"sage-backend/internal/admin/domain"
)

type AdminDBModel struct {
	ID             string     `db:"id"`
	Email          string     `db:"email"`
	PasswordHash   string     `db:"password_hash"`
	Role           string     `db:"role"`
	Status         string     `db:"status"`
	FailedAttempts int        `db:"failed_attempts"`
	LockedUntil    *time.Time `db:"locked_until"`
	LastLoginAt    *time.Time `db:"last_login_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

func (m *AdminDBModel) ToDomain() (*domain.Admin, error) {
	role, err := domain.ParseRole(m.Role)
	if err != nil {
		role = domain.RoleSupportAdmin
	}

	status, err := domain.ParseStatus(m.Status)
	if err != nil {
		status = domain.StatusActive
	}

	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	return domain.ReconstituteAdmin(
		m.ID,
		email,
		m.PasswordHash,
		role,
		status,
		m.FailedAttempts,
		m.LockedUntil,
		m.LastLoginAt,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}

func fromDomain(a *domain.Admin) *AdminDBModel {
	return &AdminDBModel{
		ID:             a.ID(),
		Email:          a.Email().String(),
		PasswordHash:   a.PasswordHash(),
		Role:           a.Role().String(),
		Status:         a.Status().String(),
		FailedAttempts: a.FailedAttempts(),
		LockedUntil:    a.LockedUntil(),
		LastLoginAt:    a.LastLoginAt(),
		CreatedAt:      a.CreatedAt(),
		UpdatedAt:      a.UpdatedAt(),
	}
}
