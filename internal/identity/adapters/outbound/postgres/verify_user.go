package postgres

import (
	"context"
)

func (r *UserRepository) MarkEmailVerified(ctx context.Context, email string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, MARK_EMAIL_VERIFIED, email)
	return err
}
