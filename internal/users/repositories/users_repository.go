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
	MarkEmailVerified(ctx context.Context, email string) error
	CreateUserWithOrganization(ctx context.Context, req *requests.CreateUserRequest, hash string) error
	GetUserOrganizations(ctx context.Context, userId string) (*[]models.Organization, error)
	Enable2FA(ctx context.Context, secret string, userID string) error
	GetTOTPSecret(ctx context.Context, userID string) (string, error)
	UpdateUserPassword(ctx context.Context, email string, hash string) error
	OnboardUserWithTransaction(ctx context.Context, req *requests.OnboardingRequest, hash string) error
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
	RETURNING id;
	`
	GET_USER_BY_EMAIL = `
	SELECT * FROM users WHERE email = $1
	`
	MARK_EMAIL_VERIFIED = `
	UPDATE users SET is_verified = true WHERE email = $1
	`
	CREATE_ORGANIZATION=`
	INSERT INTO organizations (
	name,
	owner_id,
	industry_id
	) VALUES ($1, $2, $3)
	RETURNING id;
	`
	ADD_USER_TO_ORGANIZATION=`
	INSERT INTO organization_members (
	organization_id,
	user_id,
	role
	) VALUES ($1, $2, $3)
	`
	GET_USER_ORGANIZATIONS = `
	SELECT 
	o.id, 
	o.name, 
	o.owner_id, 
	COALESCE(i.name, '') AS industry, 
	o.created_at, 
	o.updated_at, 
	o.deleted_at 
	FROM organizations o
	JOIN organization_members om ON o.id = om.organization_id
	LEFT JOIN industries i ON o.industry_id = i.id
	WHERE om.user_id = $1
	`
	ENABLE_2FA=`
		UPDATE users 
		SET 
			two_factor_enabled = true,
			two_factor_secret = $1
		WHERE id = $2`
	GET_TOTP_SECRET=`
	SELECT two_factor_secret FROM users 
	WHERE id = $1 
	AND two_factor_enabled = true
	AND is_verified = true
	`
	UPDATE_USER_PASSWORD=`
	UPDATE users SET password_hash = $1 WHERE email = $2
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
	err = tx.QueryRowContext(ctx, CREATE_ORGANIZATION, req.FirstName + " 's Org", userId).Scan(&orgId)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, userId, "owner")
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
func (r *UsersRepository) OnboardUserWithTransaction(ctx context.Context, req *requests.OnboardingRequest, hash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userId string
	err = tx.QueryRowContext(ctx, CREATE_USER, req.FirstName, req.LastName, req.Email, req.TimeZone, hash).Scan(&userId)
	if err != nil {
		return err
	}

	var orgId string
	err = tx.QueryRowContext(ctx, CREATE_ORGANIZATION, req.CompanyName, userId, req.IndustryId).Scan(&orgId)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, userId, "owner")
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}