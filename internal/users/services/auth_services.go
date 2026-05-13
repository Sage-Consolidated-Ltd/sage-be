package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"sage-backend/pkg/jwt"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
)

type AuthServiceInt interface {
	CreateUser(ctx context.Context, req *requests.CreateUserRequest) error
	OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, error)
	Login(ctx context.Context, req *requests.LoginRequest) (*models.GetUserResponse, error)
	FetchExternalUser(ctx context.Context, provider, accessToken string) (*models.ExternalUser, error)
	Generate2FA(ctx context.Context, email string) (string, string, error)
	Enabled2FA(ctx context.Context, code string, secret string, userID string) error
	Verify2FA(ctx context.Context, code, userID string) error
	ForgotPassword(ctx context.Context, email string) error
	VerifyResetToken(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, req *requests.ResetPasswordRequest, token string) error
	CreateUserWithOrganization(ctx context.Context, req *requests.OnboardingRequest) (*models.GetUserResponse, error)
	SendEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
}

type AuthService struct {
	userRepo    repositories.UsersRepositoryInt
	jwtService  *jwt.JwtService
	appConfig   *config.APIConfig
	redis       *redis.Client
	companyRepo repositories.CompanyRepositoryInt
	mailer      mailer.EmailClientInt
}

func NewAuthService(
	userRepo repositories.UsersRepositoryInt,
	jwtService *jwt.JwtService,
	appConfig *config.APIConfig,
	redis *redis.Client,
	companyRepo repositories.CompanyRepositoryInt,
	mailer mailer.EmailClientInt,
) AuthServiceInt {
	return &AuthService{
		userRepo:    userRepo,
		jwtService:  jwtService,
		appConfig:   appConfig,
		redis:       redis,
		companyRepo: companyRepo,
		mailer:      mailer,
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
	err = s.userRepo.CreateUserWithOrganization(ctx, req, hash)
	if err != nil {
		return err
	}

	// NOTE - should send mail to user for verification of email
	return nil
}
func (s *AuthService) CreateUserWithOrganization(ctx context.Context, req *requests.OnboardingRequest) (*models.GetUserResponse, error) {
	// check if email already exists
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if user != nil {
		return nil, apperrors.ConflictError("EMAIL ALREADY EXISTS")
	}
	if err != nil {
		var appErr *apperrors.ErrorResponse
		if !errors.As(err, &appErr) {
			return nil, err
		}
	}

	// check if industry is valid
	_, err = s.companyRepo.GetIndustryByID(ctx, req.IndustryId)
	if err != nil {
		return nil, err
	}

	// hash password
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("Error hashing password: %s", err)
	}

	// create user and organization
	user, err = s.userRepo.OnboardUserWithTransaction(ctx, req, hash)
	if err != nil {
		return nil, err
	}

	orgs, err := s.userRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := user.ToResponse(orgs)

	return resp, nil
}
func (s *AuthService) OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	// if err != nil {
	// 	var appErr *apperrors.ErrorResponse
	// 	if !errors.As(err, &appErr) || appErr.Code != apperrors.ErrNotFound {
	// 		return nil, err
	// 	}

	// 	password, err := utils.GenerateRandomStringForHashing(32)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	hash, err := utils.HashPassword(password)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	newUser := &requests.CreateUserRequest{
	// 		FirstName: payload.FirstName,
	// 		LastName:  payload.LastName,
	// 		Email:     payload.Email,
	// 		Password:  hash,
	// 	}

	// 	if err = s.userRepo.CreateUser(ctx, newUser, hash); err != nil {
	// 		return nil, err
	// 	}
	// 	if err = s.userRepo.MarkEmailVerified(ctx, newUser.Email); err != nil {
	// 		return nil, err
	// 	}

	// 	// Fetch the newly created user
	// 	user, err = s.userRepo.GetUserByEmail(ctx, payload.Email)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	if user == nil {
		return nil, apperrors.NotFoundError("USER NOT FOUND")
	}
	if !user.IsVerified {
		err = s.userRepo.MarkEmailVerified(ctx, payload.Email)
		if err != nil {
			return nil, err
		}
	}

	orgs, err := s.userRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := user.ToResponse(orgs)

	return resp, nil
}
func (s *AuthService) Login(ctx context.Context, req *requests.LoginRequest) (*models.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// if !user.IsVerified {
	// 	return nil, "", apperrors.UnauthorizedException("email not verified")
	// }

	if match := utils.CompareHashAndPassword(req.Password, user.PasswordHash); !match {
		return nil, apperrors.UnauthorizedException("invalid credentials")
	}

	orgs, err := s.userRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := user.ToResponse(orgs)

	return resp, nil
}
func (s *AuthService) FetchExternalUser(ctx context.Context, provider, accessToken string) (*models.ExternalUser, error) {
	var url string
	switch provider {
	case "google":
		url = "https://www.googleapis.com/oauth2/v3/userinfo"
	case "github":
		url = "https://api.github.com/user"
	case "azure":
		url = "https://graph.microsoft.com/v1.0/me"
	default:
		return nil, errors.New("unsupported provider")
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	user := &models.ExternalUser{}
	switch provider {
	case "github":
		user.ID = fmt.Sprintf("%v", raw["id"])
		user.FirstName = fmt.Sprintf("%v", raw["name"])
		if raw["email"] != nil {
			user.Email = fmt.Sprintf("%v", raw["email"])
		}
	case "google":
		user.ID = fmt.Sprintf("%v", raw["sub"])
		user.Email = fmt.Sprintf("%v", raw["email"])
		user.FirstName = fmt.Sprintf("%v", raw["given_name"])
		user.LastName = fmt.Sprintf("%v", raw["family_name"])
	}

	if user.Email == "" && provider == "github" {
		user.Email, _ = s.fetchGithubEmail(accessToken)
	}

	if user.Email == "" {
		return nil, errors.New("email is required but was not provided by provider")
	}

	return user, nil
}
func (s *AuthService) fetchGithubEmail(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no verified email found on github account")
}
func (s *AuthService) Generate2FA(ctx context.Context, email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Sage",
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}

	fa_secret, err := utils.Encrypt(key.Secret(), []byte(s.appConfig.AppEncryptionKey))
	if err != nil {
		return "", "", err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	qr := base64.StdEncoding.EncodeToString(buf.Bytes())

	return fa_secret, qr, nil
}
func (s *AuthService) Enabled2FA(ctx context.Context, code string, secret string, userID string) error {
	decryptedSecret, err := utils.Decrypt(secret, []byte(s.appConfig.AppEncryptionKey))
	if err != nil {
		return fmt.Errorf("Error decrypting secret: %v", err)
	}
	if !totp.Validate(code, decryptedSecret) {
		return fmt.Errorf("Invalid code")
	}

	if err := s.userRepo.Enable2FA(ctx, secret, userID); err != nil {
		return fmt.Errorf("Error enabling 2FA: %v", err)
	}

	return nil
}
func (s *AuthService) Verify2FA(ctx context.Context, code, userID string) error {
	secret, err := s.userRepo.GetTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}

	decryptedSecret, err := utils.Decrypt(secret, []byte(s.appConfig.AppEncryptionKey))
	if err != nil {
		return fmt.Errorf("Error decrypting secret: %v", err)
	}

	if !totp.Validate(code, decryptedSecret) {
		return fmt.Errorf("Invalid 2FA code")
	}

	return nil
}
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// generate password reset token
	token := utils.GenerateSecureOTP()
	// save token in redis with expiry
	err := s.redis.Set(ctx, "password_reset:"+token, email, 15*60*time.Second).Err()
	if err != nil {
		return err
	}
	// send email to user with reset link containing token
	// NOTE - should send mail to user for password reset
	log.Println("Password reset token for ", email, " : ", token)
	return nil
}
func (s *AuthService) VerifyResetToken(ctx context.Context, token string) error {
	// check token in redis
	_, err := s.redis.Get(ctx, "password_reset:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return apperrors.BadException("invalid or expired token")
		}
		return err
	}
	return nil
}
func (s *AuthService) ResetPassword(ctx context.Context, req *requests.ResetPasswordRequest, token string) error {
	// confirm passwords match
	if req.Password != req.ConfirmPassword {
		return apperrors.BadException("passwords do not match")
	}

	// verify confirm password change token
	email, err := s.redis.Get(ctx, "password_reset:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return apperrors.BadException("invalid or expired token")
		}
		return err
	}

	_, err = s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// hash new password
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// store new password hash in db
	err = s.userRepo.UpdateUserPassword(ctx, email, hash)
	if err != nil {
		return err
	}

	// delete tokens from redis
	s.redis.Del(ctx, "password_reset:"+token)

	return nil
}
func (s *AuthService) SendEmailVerification(ctx context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.IsVerified {
		return apperrors.BadException("email is already verified")
	}

	// generate email verification token
	token := utils.GenerateSecureOTP()
	// save token in redis with expiry
	err = s.redis.Set(ctx, "email_verification:"+token, email, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	// NOTE - should send mail to user for email verification
	if err := s.mailer.SendVerificationEmail([]string{email}, mailer.VerificationEmailData{
		Name:      user.FirstName + " " + user.LastName,
		OTP:       token,
		ExpiresIn: "24 hours",
	}); err != nil {
		return err
	}
	log.Println("Email verification token for ", email, " : ", token)
	return nil
}
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	email, err := s.redis.Get(ctx, "email_verification:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return apperrors.BadException("invalid or expired token")
		}
		return err
	}

	err = s.userRepo.MarkEmailVerified(ctx, email)
	if err != nil {
		return err
	}

	s.redis.Del(ctx, "email_verification:"+token)

	return nil
}
