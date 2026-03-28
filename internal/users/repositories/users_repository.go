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
}

var (
	CREATE_USER = `
	INSERT INTO users (
	first_name,
	last_name,
	email,
	password_hash
	) VALUES ($1, $2, $3, $4)
	`
	GET_USER_BY_EMAIL = `
	SELECT * FROM users WHERE email = $1
	`
	MARK_EMAIL_VERIFIED = `
	UPDATE users SET is_verified = true WHERE email = $1
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
