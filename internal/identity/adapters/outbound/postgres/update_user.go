package postgres

import (
	"context"
	"sage-backend/internal/identity/usecase/dto"
)

func (r *UserRepository) UpdateUser(ctx context.Context, id string, req *dto.UpdateProfileRequest) error {
	_, err := r.Executor(ctx).ExecContext(ctx, UPDATE_USER, req.FirstName, req.LastName, req.TimeZone, id)
	return err
}

func (r *UserRepository) UpdateUserContactInfo(ctx context.Context, id string, phoneNumber, backupEmail string) error {
	_, err := r.Executor(ctx).ExecContext(ctx, UPDATE_USER_CONTACT_INFO, phoneNumber, backupEmail, id)
	return err
}
