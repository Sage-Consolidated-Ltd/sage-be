package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase/dto"
)

type LoginUser struct {
	userRepo outbound.UserRepository
}

func NewLoginUser(userRepo outbound.UserRepository) *LoginUser {
	return &LoginUser{
		userRepo: userRepo,
	}
}

func (l *LoginUser) Login(ctx context.Context, req *dto.LoginRequest) (*dto.GetUserResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	user, err := l.userRepo.GetUserByEmail(ctx, email.String())
	if err != nil {
		return nil, err
	}

	if match := utils.CompareHashAndPassword(req.Password, user.PasswordHash().String()); !match {
		return nil, apperrors.UnauthorizedException("invalid credentials")
	}

	orgs, err := l.userRepo.GetUserOrganizations(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return dto.UserToResponse(user, orgs), nil
}

func (l *LoginUser) OAuthLogin(ctx context.Context, payload *dto.CreateUserRequest) (*dto.GetUserResponse, error) {
	email, err := domain.NewEmail(payload.Email)
	if err != nil {
		return nil, err
	}

	user, err := l.userRepo.GetUserByEmail(ctx, email.String())
	if user == nil {
		return nil, apperrors.NotFoundError("USER NOT FOUND")
	}
	if !user.IsVerified() {
		err = l.userRepo.MarkEmailVerified(ctx, email.String())
		if err != nil {
			return nil, err
		}
	}

	orgs, err := l.userRepo.GetUserOrganizations(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return dto.UserToResponse(user, orgs), nil
}

func (l *LoginUser) FetchExternalUser(ctx context.Context, provider, accessToken string) (*domain.ExternalUser, error) {
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

	user := &domain.ExternalUser{}
	var emailStr string
	switch provider {
	case "github":
		user.ID = fmt.Sprintf("%v", raw["id"])
		user.FirstName = fmt.Sprintf("%v", raw["name"])
		if raw["email"] != nil {
			emailStr = fmt.Sprintf("%v", raw["email"])
		}
	case "google":
		user.ID = fmt.Sprintf("%v", raw["sub"])
		emailStr = fmt.Sprintf("%v", raw["email"])
		user.FirstName = fmt.Sprintf("%v", raw["given_name"])
		user.LastName = fmt.Sprintf("%v", raw["family_name"])
	}

	if emailStr == "" && provider == "github" {
		emailStr, _ = l.fetchGithubEmail(accessToken)
	}

	if emailStr == "" {
		return nil, errors.New("email is required but was not provided by provider")
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	user.Email = email

	return user, nil
}

func (l *LoginUser) fetchGithubEmail(accessToken string) (string, error) {
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
