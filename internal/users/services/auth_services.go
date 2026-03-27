package services

import (
	"context"
	"errors"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"sage-backend/pkg/jwt"
)

type AuthServiceInt interface {
	CreateUser(ctx context.Context, req *requests.CreateUserRequest) error
}

type AuthService struct {
	userRepo   repositories.UsersRepositoryInt
	jwtService *jwt.JwtService
}

func NewAuthService(userRepo repositories.UsersRepositoryInt, jwtService *jwt.JwtService) AuthServiceInt {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, req *requests.CreateUserRequest) error {
	// check if email already exists
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if user != nil {
		return apperrors.ConflictError("EMAIL ALREADY EXISTS")
	}
	if err != nil {
		var appErr *apperrors.ErrorResponse
		if !errors.As(err, &appErr) {
			return err
		}
	}

	// hash password
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("Error hashing password: %s", err)
	}

	// then create account
	err = s.userRepo.CreateUser(ctx, req, hash)
	if err != nil {
		return err
	}

	// NOTE - should send mail to user for verification of email
	return nil
}
