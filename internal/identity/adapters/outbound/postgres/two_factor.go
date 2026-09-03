package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sage-backend/internal/shared/errors/apperrors"
)

func (r *UserRepository) Enable2FA(ctx context.Context, secret string, userID string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, ENABLE_2FA, secret, userID)
	return err
}

func (r *UserRepository) GetTOTPSecret(ctx context.Context, userID string) (string, error) {
	var secret string
	err := r.Executor(ctx).QueryRowxContext(ctx, GET_TOTP_SECRET, userID).Scan(&secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperrors.NotFoundError("2FA Secret not found")
		}
		return "", err
	}
	return secret, nil
}
