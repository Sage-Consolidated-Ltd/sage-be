package handlers

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type AuthHandler struct {
	authServ services.AuthServiceInt
	oAuthConfig *config.OAuthConfig
	appConfig *config.APIConfig
}

func NewAuthHandler(authServ services.AuthServiceInt, oAuthConfig *config.OAuthConfig, appConfig *config.APIConfig) *AuthHandler {
	return &AuthHandler{
		authServ: authServ,
		oAuthConfig: oAuthConfig,
		appConfig: appConfig,
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
	authConf := a.oAuthConfig.GetConfig(providerName)

	if authConf == nil {
        return response.Error(c, fiber.StatusBadRequest, "Unsupported auth provider", nil)
    }
	
	state, err := utils.GenerateRandomStringForHashing(32)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	c.Cookie(&fiber.Cookie{
		Name: "oauth_state",
		Value: state,
		Expires:  time.Now().Add(15 * time.Minute),
        HTTPOnly: true,
        Secure:   a.appConfig.APP_ENV == "production",
        SameSite: "Lax",
	})
	loginUrl := authConf.AuthCodeURL(state)

	return c.Redirect(loginUrl, fiber.StatusTemporaryRedirect)
}
func (a *AuthHandler) AuthCallback(c *fiber.Ctx) error {
	ctx := c.Context()
	providerName := c.Params("provider")
	queryState := c.Query("state")
	code := c.Query("code")

	cookieState := c.Cookies("oauth_state")
	if queryState == "" || queryState != cookieState {
		return response.Error(c, fiber.StatusUnauthorized, "invalid oauth state", nil)
	}

	c.ClearCookie("oauth_state")

	authConf := a.oAuthConfig.GetConfig(providerName)
	if authConf == nil {
		return response.Error(c, fiber.StatusBadRequest, "unsupported provider", nil)
	}

	token, err := authConf.Exchange(ctx, code)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to exchange token", nil)
	}

	externalUser, err := a.authServ.FetchExternalUser(ctx, providerName, token.AccessToken)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess, _ := config.Store.Get(c)

	resp, err := a.authServ.OAuthLogin(ctx, &requests.CreateUserRequest{
		FirstName: externalUser.FirstName,
		LastName:  externalUser.LastName,
		Email:     externalUser.Email,
	})

	if err != nil {
		if appErr, ok := err.(*apperrors.ErrorResponse); ok{
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}

		return response.Error(c, fiber.StatusInternalServerError, "internal server error", err.Error())
	}

	if resp.TwoFactorEnabled {
		sess.Set("userID", resp.ID)
		sess.Set("pending_2fa", true)
		if err := sess.Save(); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
		}
		return response.JSON(c, fiber.StatusAccepted, "2FA required", nil)
	}

	if _, err := config.SetSession(c, config.SessionParam{
		ID: resp.ID,
		Role: string(resp.Role),
		Email: resp.Email,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Login successful", map[string]interface{}{"data": resp})
}
func (a *AuthHandler) Login(c *fiber.Ctx) error {
	var req requests.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	sess, _ := config.Store.Get(c)
	if sess.Get("userID") != nil && sess.Get("pending_2fa") == nil {
		return response.JSON(c, fiber.StatusOK, "Login successful", nil)
	}
	
	resp, err := a.authServ.Login(c.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.ErrorResponse); ok{
			log.Errorf("Error occured with login: %s", appErr)
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		log.Errorf("Error occured with login: %s", err)
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	if resp.TwoFactorEnabled {
		sess.Set("userID", resp.ID)
		sess.Set("pending_2fa", true)
		if err := sess.Save(); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
		}
		return response.JSON(c, fiber.StatusAccepted, "2FA required", nil)
	}

	if _, err := config.SetSession(c, config.SessionParam{
		ID: resp.ID,
		Role: string(resp.Role),
		Email: resp.Email,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Login successful", map[string]interface{}{"data": resp})
}
func (a *AuthHandler) Generate2FA(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
    if err != nil || sess.Get("userID") == nil {
        return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
    }

	email := sess.Get("email").(string) 
	
	fa_secret, qrCode, err := a.authServ.Generate2FA(c.Context(), email)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Set("pending_totp_secret", fa_secret)
	sess.Save()

	return response.JSON(c, fiber.StatusOK, "scan QR with Google Authenticator", map[string]interface{}{
        "qr_code": "data:image/png;base64," + qrCode,
    })
}
func (a *AuthHandler) Enable2FA(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil || sess.Get("userID") == nil {
        return response.Error(c, fiber.StatusUnauthorized, "not authenticated", nil)
    }
	userID := sess.Get("userID").(string)

	var req requests.GoogleAuthenticatorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	secret, ok := sess.Get("pending_totp_secret").(string)
    if !ok || secret == "" {
        return response.Error(c, fiber.StatusBadRequest, "no pending 2FA setup", nil)
    }

	err = a.authServ.Enabled2FA(c.Context(), req.Code, secret, userID)
	if err != nil {
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Delete("pending_totp_secret")
    sess.Save()

	return response.JSON(c, fiber.StatusOK, "2FA Enabled", nil)
}
func (a *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil || sess.Get("pending_2fa") == nil {
        return response.Error(c, fiber.StatusUnauthorized, "not authenticated", nil)
    }

	var req requests.GoogleAuthenticatorRequest
    if err := c.BodyParser(&req); err != nil {
        return response.Error(c, fiber.StatusBadRequest, "invalid body", nil)
    }

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

    userID := sess.Get("userID").(string)

	err = a.authServ.Verify2FA(c.Context(), req.Code, userID)
	if err != nil {
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Delete("pending_2fa")
    sess.Set("verified", true)
    sess.Save()

	return response.JSON(c, fiber.StatusOK, "2FA Verified", nil)
}