package postgres

import (
	"context"
)

func (r *UserRepository) UpdateUserPassword(ctx context.Context, email string, hash string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, UPDATE_USER_PASSWORD, hash, email)
	return err
}

func (r *UserRepository) UpdatePasswordHashByID(ctx context.Context, id string, hash string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, UPDATE_USER_PASSWORD_BY_ID, hash, id)
	return err
}

