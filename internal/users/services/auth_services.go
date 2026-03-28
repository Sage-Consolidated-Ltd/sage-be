package services

import (
	"context"
	"errors"
	"fmt"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"sage-backend/pkg/jwt"
)

type AuthServiceInt interface {
	CreateUser(ctx context.Context, req *requests.CreateUserRequest) error
	OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, string, error)
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
func (s *AuthService) OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		var appErr *apperrors.ErrorResponse
		if !errors.As(err, &appErr) || appErr.Code != apperrors.ErrNotFound {
			return nil, "", err
		}

		password, err := utils.GenerateRandomStringForHashing(32)
		if err != nil {
			return nil, "", err
		}
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, "", err
		}

		newUser := &requests.CreateUserRequest{
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Email:     payload.Email,
			Password:  hash,
		}

		if err = s.userRepo.CreateUser(ctx, newUser, hash); err != nil {
			return nil, "", err
		}
		if err = s.userRepo.MarkEmailVerified(ctx, newUser.Email); err != nil {
			return nil, "", err
		}

		// Fetch the newly created user
		user, err = s.userRepo.GetUserByEmail(ctx, payload.Email)
		if err != nil {
			return nil, "", err
		}
	}
	if !user.IsVerified {
		err = s.userRepo.MarkEmailVerified(ctx, payload.Email)
		if err != nil {
			return nil, "", err
		}
	}

	token, err := s.jwtService.GenerateToken(jwt.UserPayload{
		Id: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	})
	if err != nil {
		return nil, "", apperrors.InternalServerError("error signing token")
	}

	resp := user.ToResponse()

	return resp, token, nil
}
