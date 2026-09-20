package postgres

import (
	"context"
)

func (r *UserRepository) SoftDeleteUser(ctx context.Context, id string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, SOFT_DELETE_USER, id)
	return err
}
