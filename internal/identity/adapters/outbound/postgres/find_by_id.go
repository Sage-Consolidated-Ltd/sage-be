package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/identity/domain"
)

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var model userModel
	err := r.Executor(ctx).GetContext(ctx, &model, GET_USER_BY_ID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("USER NOT FOUND")
		}
		return nil, err
	}
	return model.ToDomain()
}
