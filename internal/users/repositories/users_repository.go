package repositories

import (
	"context"
	"database/sql"
	"errors"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/requests"
)

type UsersRepositoryInt interface {
	CreateUser(ctx context.Context, req *requests.CreateUserRequest, hash string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	MarkEmailVerified(ctx context.Context, email string) error
	CreateUserWithOrganization(ctx context.Context, req *requests.CreateUserRequest, hash string) error
	GetUserOrganizations(ctx context.Context, userId string) (*[]models.Organization, error)
	Enable2FA(ctx context.Context, secret string, userID string) error
	Reset2FA(ctx context.Context, userID string) error
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	UpdateUserPassword(ctx context.Context, email string, hash string) error
	OnboardUserWithTransaction(ctx context.Context, req *requests.OnboardingRequest, hash string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, req *requests.UpdateProfileRequest) error
}

var (
	CREATE_USER = `
	INSERT INTO users (
	first_name,
	last_name,
	email,
	time_zone,
	password_hash
	) VALUES ($1, $2, $3, $4, $5)
	RETURNING *;
	`
	GET_USER_BY_EMAIL = `
	SELECT * FROM users WHERE email = $1
	`
	GET_USER_BY_ID = `
	SELECT * FROM users WHERE id = $1
	`
	MARK_EMAIL_VERIFIED = `
	UPDATE users SET is_verified = true WHERE email = $1
	`
	CREATE_ORGANIZATION = `
	INSERT INTO organizations (
	name,
	owner_id,
	industry_id
	) VALUES ($1, $2, $3)
	RETURNING id;
	`
	ADD_USER_TO_ORGANIZATION = `
	INSERT INTO organization_members (
	organization_id,
	user_id,
	role_id, 
	status
	) VALUES ($1, $2, $3, $4)
	`
	GET_USER_ORGANIZATIONS = `
	SELECT 
	o.id, 
	o.name, 
	o.owner_id, 
	COALESCE(i.name, '') AS industry,
	COALESCE(r.name, '') AS role, 
	om.status,
	o.created_at, 
	o.updated_at, 
	o.deleted_at 
	FROM organizations o
	JOIN organization_members om ON o.id = om.organization_id
	LEFT JOIN industries i ON o.industry_id = i.id
	LEFT JOIN organization_roles r ON om.role_id = r.id
	WHERE om.user_id = $1
	`
	ENABLE_2FA = `
		UPDATE users 
		SET 
			two_factor_enabled = true,
			two_factor_secret = $1
		WHERE id = $2`
	RESET_2FA = `
		UPDATE users
		SET
			two_factor_enabled = false,
			two_factor_secret = NULL
		WHERE id = $1`
	GET_TOTP_SECRET = `
	SELECT two_factor_secret FROM users 
	WHERE id = $1 
	AND two_factor_enabled = true
	AND is_verified = true
	`
	UPDATE_USER_PASSWORD = `
	UPDATE users SET password_hash = $1 WHERE email = $2
	`
	UPDATE_USER = `
	UPDATE users SET 
		first_name = $1,
		last_name = $2,
		time_zone = $3,
		updated_at = NOW()
	WHERE id = $4
	`
	GET_ORGANIZATION_ROLE_ID = `
	SELECT id FROM organization_roles WHERE name = $1`
	INVITE_MEMBER_TO_ORGANIZATION = `
	INSERT INTO organization_invites (
		organization_id,
		email,
		role_id,
		invited_by,
		token_hash,
		expires_at
	)
	SELECT $1, LOWER($2), $3, $4, $5, $6
	WHERE NOT EXISTS (
		SELECT 1 FROM organization_invites oi
		WHERE oi.organization_id = $1
		AND LOWER(oi.email) = LOWER($2)
		AND oi.status = 'pending'
	)
	ON CONFLICT DO NOTHING
	RETURNING id;
	`
)

type UsersRepository struct {
	db *db.DB
}

func NewUsersRepository(db *db.DB) UsersRepositoryInt {
	return &UsersRepository{
		db: db,
	}
}

func (r *UsersRepository) CreateUser(ctx context.Context, req *requests.CreateUserRequest, hash string) error {
	_, err := r.db.ExecContext(
		ctx,
		CREATE_USER,
		req.FirstName,
		req.LastName,
		req.Email,
		hash,
	)
	if err != nil {
		return err
	}
	return nil
}
func (r *UsersRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.db.GetContext(ctx, &user, GET_USER_BY_EMAIL, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("USER NOT FOUND")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UsersRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User

	err := r.db.GetContext(ctx, &user, GET_USER_BY_ID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("USER NOT FOUND")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UsersRepository) UpdateUser(ctx context.Context, id string, req *requests.UpdateProfileRequest) error {
	_, err := r.db.ExecContext(ctx, UPDATE_USER, req.FirstName, req.LastName, req.TimeZone, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) MarkEmailVerified(ctx context.Context, email string) error {
	_, err := r.db.ExecContext(ctx, MARK_EMAIL_VERIFIED, email)
	if err != nil {
		return err
	}
	return nil
}
func (r *UsersRepository) CreateUserWithOrganization(ctx context.Context, req *requests.CreateUserRequest, hash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var userId string
	err = tx.QueryRowContext(ctx, CREATE_USER, req.FirstName, req.LastName, req.Email, hash).Scan(&userId)
	if err != nil {
		return err
	}

	var orgId string
	err = tx.QueryRowContext(ctx, CREATE_ORGANIZATION, req.FirstName+" 's Org", userId).Scan(&orgId)
	if err != nil {
		return err
	}

	var roleId string
	if err = tx.QueryRowContext(ctx, GET_ORGANIZATION_ROLE_ID, "owner").Scan(&roleId); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, userId, roleId)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
func (r *UsersRepository) GetUserOrganizations(ctx context.Context, userId string) (*[]models.Organization, error) {
	var orgs []models.Organization

	err := r.db.SelectContext(ctx, &orgs, GET_USER_ORGANIZATIONS, userId)
	if err != nil {
		return nil, err
	}
	return &orgs, nil
}
func (r *UsersRepository) Enable2FA(ctx context.Context, secret string, userID string) error {
	_, err := r.db.ExecContext(ctx, ENABLE_2FA, secret, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *UsersRepository) Reset2FA(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, RESET_2FA, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *UsersRepository) GetTOTPSecret(ctx context.Context, userID string) (string, error) {
	var secret string
	err := r.db.QueryRowContext(ctx, GET_TOTP_SECRET, userID).Scan(&secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperrors.NotFoundError("2FA Secret not found")
		}
		return "", err
	}
	return secret, nil
}
func (r *UsersRepository) UpdateUserPassword(ctx context.Context, email string, hash string) error {
	_, err := r.db.ExecContext(ctx, UPDATE_USER_PASSWORD, hash, email)
	if err != nil {
		return err
	}
	return nil
}
func (r *UsersRepository) OnboardUserWithTransaction(ctx context.Context, req *requests.OnboardingRequest, hash string) (*models.User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user models.User
	err = tx.QueryRowxContext(ctx, CREATE_USER, req.FirstName, req.LastName, req.Email, req.TimeZone, hash).StructScan(&user)
	if err != nil {
		return nil, err
	}

	var orgId string
	err = tx.QueryRowContext(ctx, CREATE_ORGANIZATION, req.CompanyName, user.ID, req.IndustryId).Scan(&orgId)
	if err != nil {
		return nil, err
	}

	var roleId string
	if err = tx.QueryRowContext(ctx, GET_ORGANIZATION_ROLE_ID, "owner").Scan(&roleId); err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, user.ID, roleId, "active")
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}
