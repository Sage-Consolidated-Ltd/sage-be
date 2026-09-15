package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/shared/db"
)

type AdminRepository struct {
	db.Repository
}

func NewAdminRepository(database *db.DB) outbound.AdminRepository {
	return &AdminRepository{
		Repository: db.NewRepository(database),
	}
}

var _ outbound.AdminRepository = (*AdminRepository)(nil)

const (
	queryGetAdminByID = `
		SELECT id, email, password_hash, role, status, failed_attempts, locked_until, last_login_at, created_at, updated_at
		FROM admins
		WHERE id = $1
	`
	queryGetAdminByEmail = `
		SELECT id, email, password_hash, role, status, failed_attempts, locked_until, last_login_at, created_at, updated_at
		FROM admins
		WHERE LOWER(email) = LOWER($1)
	`
	queryCreateAdmin = `
		INSERT INTO admins (id, email, password_hash, role, status, failed_attempts, locked_until, last_login_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	queryUpdateAdmin = `
		UPDATE admins
		SET email = $2, password_hash = $3, role = $4, status = $5, failed_attempts = $6, locked_until = $7, last_login_at = $8, updated_at = $9
		WHERE id = $1
	`
)

func (r *AdminRepository) GetByID(ctx context.Context, id string) (*domain.Admin, error) {
	var model AdminDBModel
	err := r.Executor(ctx).GetContext(ctx, &model, queryGetAdminByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var model AdminDBModel
	err := r.Executor(ctx).GetContext(ctx, &model, queryGetAdminByEmail, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *AdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	m := fromDomain(admin)
	_, err := r.Executor(ctx).ExecContext(
		ctx,
		queryCreateAdmin,
		m.ID,
		m.Email,
		m.PasswordHash,
		m.Role,
		m.Status,
		m.FailedAttempts,
		m.LockedUntil,
		m.LastLoginAt,
		m.CreatedAt,
		m.UpdatedAt,
	)
	return err
}

func (r *AdminRepository) Update(ctx context.Context, admin *domain.Admin) error {
	m := fromDomain(admin)
	_, err := r.Executor(ctx).ExecContext(
		ctx,
		queryUpdateAdmin,
		m.ID,
		m.Email,
		m.PasswordHash,
		m.Role,
		m.Status,
		m.FailedAttempts,
		m.LockedUntil,
		m.LastLoginAt,
		m.UpdatedAt,
	)
	return err
}
