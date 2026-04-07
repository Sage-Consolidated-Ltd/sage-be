package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"sage-backend/pkg/jwt"

	"github.com/pquerna/otp/totp"
)

type AuthServiceInt interface {
	CreateUser(ctx context.Context, req *requests.CreateUserRequest) error
	OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, error)
	Login(ctx context.Context, req *requests.LoginRequest) (*models.GetUserResponse, error)
	FetchExternalUser(ctx context.Context, provider, accessToken string) (*models.ExternalUser, error)
	Generate2FA(ctx context.Context, email string) (string, string, error)
	Enabled2FA(ctx context.Context, code string, secret string, userID string) error
	Verify2FA(ctx context.Context, code, userID string) error
}

type AuthService struct {
	userRepo   repositories.UsersRepositoryInt
	jwtService *jwt.JwtService
	appConfig *config.Config
}

func NewAuthService(userRepo repositories.UsersRepositoryInt, jwtService *jwt.JwtService, appConfig *config.Config) AuthServiceInt {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		appConfig: appConfig,
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
func (s *AuthService) OAuthLogin(ctx context.Context, payload *requests.CreateUserRequest) (*models.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		var appErr *apperrors.ErrorResponse
		if !errors.As(err, &appErr) || appErr.Code != apperrors.ErrNotFound {
			return nil, err
		}

		password, err := utils.GenerateRandomStringForHashing(32)
		if err != nil {
			return nil, err
		}
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}

		newUser := &requests.CreateUserRequest{
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Email:     payload.Email,
			Password:  hash,
		}

		if err = s.userRepo.CreateUser(ctx, newUser, hash); err != nil {
			return nil, err
		}
		if err = s.userRepo.MarkEmailVerified(ctx, newUser.Email); err != nil {
			return nil, err
		}

		// Fetch the newly created user
		user, err = s.userRepo.GetUserByEmail(ctx, payload.Email)
		if err != nil {
			return nil, err
		}
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
		Issuer: "Sage",
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