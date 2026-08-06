package postgres

import (
	"context"
	"sage-backend/internal/identity/usecase/dto"
)

func (r *UserRepository) CreateUser(ctx context.Context, req *dto.CreateUserRequest, hash string) error {
	_, err := r.Executor(ctx).ExecContext(
		ctx,
		CREATE_USER,
		req.FirstName,
		req.LastName,
		req.Email,
		hash,
	)
	return err
}
