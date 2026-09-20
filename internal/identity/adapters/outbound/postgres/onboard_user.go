package postgres

import (
	"context"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/usecase/dto"
)

func (r *UserRepository) CreateUserWithOrganization(ctx context.Context, req *dto.CreateUserRequest, hash string) error {
	var userId string
	err := r.Executor(ctx).QueryRowxContext(ctx, CREATE_USER, req.FirstName, req.LastName, req.Email, hash).Scan(&userId)
	if err != nil {
		return err
	}

	var orgId string
	err = r.Executor(ctx).QueryRowxContext(ctx, CREATE_ORGANIZATION, req.FirstName+" 's Org", userId, nil).Scan(&orgId)
	if err != nil {
		return err
	}

	var roleId string
	if err = r.Executor(ctx).QueryRowxContext(ctx, GET_ORGANIZATION_ROLE_ID, "owner").Scan(&roleId); err != nil {
		return err
	}

	_, err = r.Executor(ctx).ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, userId, roleId)
	return err
}

func (r *UserRepository) OnboardUserWithTransaction(ctx context.Context, req *dto.OnboardingRequest, hash string) (*domain.User, error) {
	var model userModel
	row := r.Executor(ctx).QueryRowxContext(ctx, CREATE_USER, req.FirstName, req.LastName, req.Email, req.TimeZone, hash)
	if err := row.StructScan(&model); err != nil {
		return nil, err
	}

	var industryID interface{}
	if req.IndustryId != "" {
		industryID = req.IndustryId
	}

	var orgId string
	err := r.Executor(ctx).QueryRowxContext(ctx, CREATE_ORGANIZATION, req.CompanyName, model.ID, industryID).Scan(&orgId)
	if err != nil {
		return nil, err
	}

	var roleId string
	if err = r.Executor(ctx).QueryRowxContext(ctx, GET_ORGANIZATION_ROLE_ID, "owner").Scan(&roleId); err != nil {
		return nil, err
	}

	_, err = r.Executor(ctx).ExecContext(ctx, ADD_USER_TO_ORGANIZATION, orgId, model.ID, roleId, "active")
	if err != nil {
		return nil, err
	}

	return model.ToDomain()
}
