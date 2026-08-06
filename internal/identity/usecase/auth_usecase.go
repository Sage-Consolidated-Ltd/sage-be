package usecase

import (
	"context"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/inbound"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase/dto"
	orgOutbound "sage-backend/internal/organization/ports/outbound"
	"sage-backend/pkg/jwt"

	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	registerUser      *RegisterUser
	loginUser         *LoginUser
	twoFactorAuth     *TwoFactorAuth
	passwordReset     *PasswordReset
	emailVerification *EmailVerification
}

func NewAuthService(
	userRepo outbound.UserRepository,
	jwtService *jwt.JwtService,
	appConfig *config.APIConfig,
	redis *redis.Client,
	companyRepo orgOutbound.CompanyRepository,
	mailer mailer.EmailClientInt,
) inbound.AuthUseCase {
	return &AuthService{
		registerUser:      NewRegisterUser(userRepo, companyRepo),
		loginUser:         NewLoginUser(userRepo),
		twoFactorAuth:     NewTwoFactorAuth(userRepo, appConfig),
		passwordReset:     NewPasswordReset(userRepo, redis),
		emailVerification: NewEmailVerification(userRepo, redis, mailer),
	}
}

func (s *AuthService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) error {
	return s.registerUser.CreateUser(ctx, req)
}

func (s *AuthService) CreateUserWithOrganization(ctx context.Context, req *dto.OnboardingRequest) (*dto.GetUserResponse, error) {
	return s.registerUser.CreateUserWithOrganization(ctx, req)
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.GetUserResponse, error) {
	return s.loginUser.Login(ctx, req)
}

func (s *AuthService) OAuthLogin(ctx context.Context, payload *dto.CreateUserRequest) (*dto.GetUserResponse, error) {
	return s.loginUser.OAuthLogin(ctx, payload)
}

func (s *AuthService) FetchExternalUser(ctx context.Context, provider, accessToken string) (*domain.ExternalUser, error) {
	return s.loginUser.FetchExternalUser(ctx, provider, accessToken)
}

func (s *AuthService) Generate2FA(ctx context.Context, email string) (string, string, error) {
	return s.twoFactorAuth.Generate2FA(ctx, email)
}

func (s *AuthService) Enabled2FA(ctx context.Context, code string, secret string, userID string) error {
	return s.twoFactorAuth.Enabled2FA(ctx, code, secret, userID)
}

func (s *AuthService) Verify2FA(ctx context.Context, code, userID string) error {
	return s.twoFactorAuth.Verify2FA(ctx, code, userID)
}


func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	return s.passwordReset.ForgotPassword(ctx, email)
}

func (s *AuthService) VerifyResetToken(ctx context.Context, token string) error {
	return s.passwordReset.VerifyResetToken(ctx, token)
}

func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest, token string) error {
	return s.passwordReset.ResetPassword(ctx, req, token)
}

func (s *AuthService) SendEmailVerification(ctx context.Context, email string) error {
	return s.emailVerification.SendEmailVerification(ctx, email)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	return s.emailVerification.VerifyEmail(ctx, token)
}
