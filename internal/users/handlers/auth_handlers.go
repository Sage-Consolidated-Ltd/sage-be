package handlers

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authServ    services.AuthServiceInt
	oAuthConfig *config.OAuthConfig
	appConfig   *config.APIConfig
}

func NewAuthHandler(authServ services.AuthServiceInt, oAuthConfig *config.OAuthConfig, appConfig *config.APIConfig) *AuthHandler {
	return &AuthHandler{
		authServ:    authServ,
		oAuthConfig: oAuthConfig,
		appConfig:   appConfig,
	}
}

func (a *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req requests.OnboardingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	sess, _ := config.Store.Get(c)

	resp, err := a.authServ.CreateUserWithOrganization(c.Context(), &req)
	if err != nil {
		logger.Error("Error with AuthHandler.CreateUser: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error", nil)
	}

	if !resp.IsVerified {
		sess.Set("userID", resp.ID)
		sess.Set("pending_email_verification", true)
		if err := sess.Save(); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
		}
	}

	var orgID string
	if len(resp.Organization) > 0 {
		for _, org := range resp.Organization {
			if org.OwnerID == resp.ID {
				orgID = org.ID
			}
		}
	}

	if _, err := config.SetSession(c, config.SessionParam{
		ID:             resp.ID,
		Role:           string(resp.Role),
		Email:          resp.Email,
		OrganizationId: orgID,
	}); err != nil {
		logger.Error("Error with AuthHandler.CreateUser: ", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
	}

	return response.JSON(c, fiber.StatusOK, "User and organization created successfully.", nil)
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
		Name:     "oauth_state",
		Value:    state,
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
		logger.Error("Error with AuthHandler.AuthCallback: ", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
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

	var orgID, roleInOrg string
	if len(resp.Organization) > 0 {
		for _, org := range resp.Organization {
			if org.OwnerID == resp.ID {
				orgID = org.ID
				roleInOrg = org.Role
			}
		}
	}

	if orgID == "" && len(resp.Organization) > 0 {
		orgID = resp.Organization[0].ID
		roleInOrg = resp.Organization[0].Role
	}

	if _, err := config.SetSession(c, config.SessionParam{
		ID:                   resp.ID,
		Role:                 string(resp.Role),
		Email:                resp.Email,
		OrganizationId:       orgID,
		ActiveOrganizationID: orgID,
		RoleInOrganization:   roleInOrg,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Login successful", resp)
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
		logger.Error("Error with AuthHandler.Login :", zap.Error(err))
		if appErr, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, appErr.StatusCode, appErr.Message, nil)
		}
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

	var orgID, roleInOrg string
	if len(resp.Organization) > 0 {
		for _, org := range resp.Organization {
			if org.OwnerID == resp.ID {
				orgID = org.ID
				roleInOrg = org.Role
			}
		}
	}
	if orgID == "" && len(resp.Organization) > 0 {
		orgID = resp.Organization[0].ID
		roleInOrg = resp.Organization[0].Role
	}

	if _, err := config.SetSession(c, config.SessionParam{
		ID:                   resp.ID,
		Role:                 string(resp.Role),
		Email:                resp.Email,
		OrganizationId:       orgID,
		ActiveOrganizationID: orgID,
		RoleInOrganization:   roleInOrg,
	}); err != nil {
		logger.Error("Error with AuthHandler.Login: ", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "error setting up session", nil)
	}

	return response.JSON(c, fiber.StatusOK, "Login successful", resp)
}
func (a *AuthHandler) Generate2FA(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil || sess.Get("userID") == nil {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	email := sess.Get("email").(string)

	fa_secret, qrCode, err := a.authServ.Generate2FA(c.Context(), email)
	if err != nil {
		logger.Error("Error with AuthHandler.Generate2FA: ", zap.Error(err))
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
		logger.Error("Error with AuthHandler.Enable2FA: ", zap.Error(err))
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
		logger.Error("Error with AuthHandler.Verify2FA: ", zap.Error(err))
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
func (a *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req requests.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	err := a.authServ.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		logger.Error("Error with AuthHandler.ForgotPassword: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.JSON(c, fiber.StatusOK, "Password reset token sent to email, if exists", nil)
}
func (a *AuthHandler) VerifyResetToken(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "not authenticated", nil)
	}
	var req requests.VerifyResetTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	err = a.authServ.VerifyResetToken(c.Context(), req.Token)
	if err != nil {
		logger.Error("Error with AuthHandler.VerifyResetToken: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}
	sess.Set("password_reset_token", req.Token)
	sess.Set("password_token_verified", true)
	sess.Save()

	return response.JSON(c, fiber.StatusOK, "Token valid", nil)
}
func (a *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "not authenticated", nil)
	}
	var req requests.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	tokenVerified := sess.Get("password_token_verified")
	if tokenVerified == nil || tokenVerified.(bool) == false {
		return response.Error(c, fiber.StatusUnauthorized, "token not verified", nil)
	}
	token := sess.Get("password_reset_token")
	if token == nil || token.(string) == "" {
		return response.Error(c, fiber.StatusUnauthorized, "invalid token", nil)
	}

	err = a.authServ.ResetPassword(c.Context(), &req, token.(string))
	if err != nil {
		logger.Error("Error with AuthHandler.ResetPassword: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Delete("password_reset_token")
	sess.Delete("password_token_verified")
	sess.Save()

	return response.JSON(c, fiber.StatusOK, "Password reset successful", nil)
}
func (a *AuthHandler) SendVerificationEmail(c *fiber.Ctx) error {
	var req requests.SendVerificationEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	err := a.authServ.SendEmailVerification(c.Context(), req.Email)
	if err != nil {
		logger.Error("Error with AuthHandler.SendVerificationEmail: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.JSON(c, fiber.StatusOK, "Verification email sent", nil)
}
func (a *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req requests.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	sess, _ := config.Store.Get(c)
	err := a.authServ.VerifyEmail(c.Context(), req.Token)
	if err != nil {
		logger.Error("Error with AuthHandler.VerifyEmail: ", zap.Error(err))
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	sess.Delete("pending_email_verification")
	sess.Save()

	return response.JSON(c, fiber.StatusOK, "Email verified successfully", nil)
}
