package inbound

import (
	"context"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/usecase/dto"
)

type AuthUseCase interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) error
	CreateUserWithOrganization(ctx context.Context, req *dto.OnboardingRequest) (*dto.GetUserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.GetUserResponse, error)
	OAuthLogin(ctx context.Context, payload *dto.CreateUserRequest) (*dto.GetUserResponse, error)
	FetchExternalUser(ctx context.Context, provider, accessToken string) (*domain.ExternalUser, error)
	Generate2FA(ctx context.Context, email string) (string, string, error)
	Enabled2FA(ctx context.Context, code string, secret string, userID string) error
	Verify2FA(ctx context.Context, code, userID string) error
	ForgotPassword(ctx context.Context, email string) error
	VerifyResetToken(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest, token string) error
	SendEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
}
