package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

type AuthHandler struct {
	authServ services.AuthServiceInt
}

func NewAuthHandler(authServ services.AuthServiceInt) *AuthHandler {
	return &AuthHandler{
		authServ: authServ,
	}
}

func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req requests.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	// create user
	if err := h.authServ.CreateUser(c.Context(), &req); err != nil {
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.JSON(c, fiber.StatusOK, "User created successfully. A mail has been forwarded to verify account", nil)
}
func (a *AuthHandler) BeginAuthLogin(c *fiber.Ctx) error {
	providerName := c.Params("provider")
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	loginUrl, err := provider.GetBeginAuthURL(nil, nil)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return c.Redirect(loginUrl, fiber.StatusTemporaryRedirect)
}
func (a *AuthHandler) AuthCallback(c *fiber.Ctx) error {
	ctx := c.Context()
	providerName := c.Params("provider")
	provider, err := gomniauth.Provider(providerName)

	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	creds, err := provider.CompleteAuth(objx.MustFromURLQuery(string(c.Request().URI().QueryString())))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	user, err := provider.GetUser(creds)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}
	email := user.Email()
	if email == "" {
		if providerName == "github" {
			email, err = fetchGithubEmail(creds.Get("access_token").Str())
			if err != nil {
				return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
			}
		} else{
			return response.Error(c, fiber.StatusBadRequest, "Email is required, set public email on github/google", nil)
		}
	}

	// login user or create new user
	resp, token, err := a.authServ.OAuthLogin(ctx, &requests.CreateUserRequest{
		FirstName: user.Name(),
		LastName: user.Name(),
		Email: email,
	})

	if err != nil {
		if appErr, ok := err.(*apperrors.ErrorResponse); ok{
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}

		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Login successful", map[string]interface{}{"data": resp, "token": token})
}

func fetchGithubEmail(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email string `json:"email"`
		Primary bool `json:"primary"`
		Verified bool `json:"verified"`
	}
	 if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
        return "", err
    }

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	return "", errors.New("no verified primary email found on GitHub account")
}