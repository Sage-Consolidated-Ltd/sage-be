package usecase

import (
	"context"
	"errors"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase/dto"
	orgOutbound "sage-backend/internal/organization/ports/outbound"
)

type RegisterUser struct {
	userRepo    outbound.UserRepository
	companyRepo orgOutbound.CompanyRepository
}

func NewRegisterUser(userRepo outbound.UserRepository, companyRepo orgOutbound.CompanyRepository) *RegisterUser {
	return &RegisterUser{
		userRepo:    userRepo,
		companyRepo: companyRepo,
	}
}

func (r *RegisterUser) CreateUser(ctx context.Context, req *dto.CreateUserRequest) error {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return err
	}

	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return err
	}

	user, err := r.userRepo.GetUserByEmail(ctx, email.String())
	if user != nil {
		return apperrors.ConflictError("EMAIL ALREADY EXISTS")
	}
	if err != nil {
		var appErr *apperrors.ErrorResponse
		if !errors.As(err, &appErr) {
			return err
		}
	}

	rawHash, err := utils.HashPassword(password.String())
	if err != nil {
		return fmt.Errorf("Error hashing password: %w", err)
	}

	hash := domain.NewPasswordHash(rawHash)

	err = r.userRepo.CreateUserWithOrganization(ctx, req, hash.String())
	if err != nil {
		return err
	}

	return nil
}

func (r *RegisterUser) CreateUserWithOrganization(ctx context.Context, req *dto.OnboardingRequest) (*dto.GetUserResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return nil, err
	}

	existingUser, err := r.userRepo.GetUserByEmail(ctx, email.String())
	if err == nil && existingUser != nil {
		return nil, apperrors.BadException("EMAIL ALREADY EXISTS")
	}

	rawHash, err := utils.HashPassword(password.String())
	if err != nil {
		return nil, fmt.Errorf("Error hashing password: %w", err)
	}

	hash := domain.NewPasswordHash(rawHash)

	user, err := r.userRepo.OnboardUserWithTransaction(ctx, req, hash.String())
	if err != nil {
		return nil, err
	}

	orgs, err := r.userRepo.GetUserOrganizations(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return dto.UserToResponse(user, orgs), nil
}
