package http

import (
	"errors"
	"strings"
	"time"

	"sage-backend/internal/admin/adapters/inbound/http/dto"
	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/inbound"
	ucdto "sage-backend/internal/admin/usecase/dto"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles HTTP requests for administrative authentication flows,
// including credential login challenge, 2FA OTP verification, OTP resend,
// profile inspection, and session termination.
type AuthHandler struct {
	authUseCase inbound.AuthUseCase
	baseConfig  *config.BaseConfig
}

// NewAuthHandler constructs a new AuthHandler with its inbound use case and base configuration.
func NewAuthHandler(authUseCase inbound.AuthUseCase, baseConfig *config.BaseConfig) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		baseConfig:  baseConfig,
	}
}

// Login handles the primary administrator authentication challenge.
// It parses and validates credentials (email and password), applies progressive
// lockout policies (30-minute lock after 5 failures, permanent lock after 10 failures),
// and triggers generation of a 6-digit OTP dispatched to the administrator's email.
//
// Responses:
//   - 200 OK: Challenge initiated, returns AdminID and user message.
//   - 400 Bad Request: Malformed JSON payload.
//   - 401 Unauthorized: Invalid credentials.
//   - 403 Forbidden: Account temporarily or permanently locked, or suspended.
//   - 422 Unprocessable Entity: Request failed schema validation rules.
//   - 500 Internal Server Error: Unexpected server or storage failure.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}

	result, err := h.authUseCase.Login(c.Context(), ucdto.AdminLoginInput{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
		ClientIP:   c.IP(),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountPermanentlyLocked):
			return response.Error(c, fiber.StatusForbidden, "Account is permanently locked. Please contact a Super Admin.", nil)
		case errors.Is(err, domain.ErrAccountTemporarilyLocked):
			return response.Error(c, fiber.StatusForbidden, "Account is temporarily locked due to repeated failed attempts. Please try again in 30 minutes.", nil)
		case errors.Is(err, domain.ErrAccountSuspended):
			return response.Error(c, fiber.StatusForbidden, "Account is suspended.", nil)
		case errors.Is(err, domain.ErrInvalidCredentials):
			return response.Error(c, fiber.StatusUnauthorized, "Invalid email or password", nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, "An unexpected error occurred", err.Error())
		}
	}

	resp := dto.AdminLoginResponse{
		AdminID: result.AdminID,
		Message: result.Message,
	}
	return response.JSON(c, fiber.StatusOK, "Login challenge initiated", resp)
}

// VerifyOTP validates the 6-digit verification code provided by the administrator.
// Upon successful verification, it provisions a distributed session in Redis (with IP
// binding for anomaly detection), writes the HTTP-only admin_session cookie, and
// returns the authenticated profile and session expiration metadata.
//
// Responses:
//   - 200 OK: Authentication successful, session cookie configured.
//   - 400 Bad Request: Invalid or expired OTP code or malformed payload.
//   - 403 Forbidden: Account locked.
//   - 422 Unprocessable Entity: Validation error on input fields.
//   - 500 Internal Server Error: Session storage or system failure.
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req dto.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}

	result, err := h.authUseCase.VerifyOTP(c.Context(), ucdto.VerifyOTPInput{
		AdminID:  req.AdminID,
		OTP:      req.OTP,
		ClientIP: c.IP(),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOTP):
			return response.Error(c, fiber.StatusBadRequest, "Invalid or expired verification code", nil)
		case errors.Is(err, domain.ErrAccountPermanentlyLocked):
			return response.Error(c, fiber.StatusForbidden, "Account is permanently locked", nil)
		case errors.Is(err, domain.ErrAccountTemporarilyLocked):
			return response.Error(c, fiber.StatusForbidden, "Account is temporarily locked", nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, "Verification failed", err.Error())
		}
	}

	cookieDomain := h.baseConfig.CookieDomain
	if cookieDomain == "localhost" || cookieDomain == "127.0.0.1" {
		cookieDomain = ""
	}

	sameSite := "None"
	if h.baseConfig.APP_ENV == "production" || h.baseConfig.APP_ENV == "test" || h.baseConfig.APP_ENV == "testing" {
		sameSite = "Lax"
	}

	maxAgeSeconds := int(time.Until(result.ExpiresAt).Seconds())
	if maxAgeSeconds < 0 {
		maxAgeSeconds = 3600
	}

	c.Cookie(&fiber.Cookie{
		Name:     AdminSessionCookie,
		Value:    result.SessionID,
		Path:     "/",
		Domain:   cookieDomain,
		MaxAge:   maxAgeSeconds,
		Expires:  result.ExpiresAt,
		Secure:   h.baseConfig.APP_ENV == "production" || h.baseConfig.APP_ENV == "staging",
		HTTPOnly: true,
		SameSite: sameSite,
	})

	resp := dto.AdminAuthResponse{
		SessionID: result.SessionID,
		ExpiresAt: result.ExpiresAt,
		Admin: dto.AdminProfileResponse{
			ID:          result.Admin.ID,
			Email:       result.Admin.Email,
			Role:        result.Admin.Role,
			Status:      result.Admin.Status,
			LastLoginAt: result.Admin.LastLoginAt,
			CreatedAt:   result.Admin.CreatedAt,
		},
	}

	return response.JSON(c, fiber.StatusOK, "Authentication successful", resp)
}

// ResendOTP triggers dispatch of a fresh 6-digit verification code to the
// administrator's email. It verifies the admin existence, lockout state, and
// replaces any active OTP challenge with a new one.
//
// Responses:
//   - 200 OK: New OTP dispatched to administrator's email.
//   - 400 Bad Request: Malformed JSON payload.
//   - 403 Forbidden: Account is temporarily or permanently locked.
//   - 404 Not Found: Admin record does not exist.
//   - 422 Unprocessable Entity: Validation error on input fields.
//   - 500 Internal Server Error: Email transport or storage failure.
func (h *AuthHandler) ResendOTP(c *fiber.Ctx) error {
	var req dto.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}

	if err := h.authUseCase.ResendOTP(c.Context(), ucdto.ResendOTPInput{
		AdminID:  req.AdminID,
		ClientIP: c.IP(),
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountPermanentlyLocked), errors.Is(err, domain.ErrAccountTemporarilyLocked):
			return response.Error(c, fiber.StatusForbidden, "Account is locked", nil)
		case errors.Is(err, domain.ErrAdminNotFound):
			return response.Error(c, fiber.StatusNotFound, "Admin not found", nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, "Failed to resend code", err.Error())
		}
	}

	return response.JSON(c, fiber.StatusOK, "New verification code dispatched", nil)
}

// Me retrieves the profile of the currently authenticated administrator.
// The caller identity is derived from the Fiber context locals populated by
// AdminAuthMiddleware (via session cookie or Authorization header).
//
// Responses:
//   - 200 OK: Admin profile returned successfully.
//   - 401 Unauthorized: Caller is missing authentication context.
//   - 404 Not Found: Admin record was deleted or not found.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	adminID, ok := c.Locals(LocalAdminID).(string)
	if !ok || adminID == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	profile, err := h.authUseCase.GetAdminProfile(c.Context(), adminID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Admin profile not found", err.Error())
	}

	resp := dto.AdminProfileResponse{
		ID:          profile.ID,
		Email:       profile.Email,
		Role:        profile.Role,
		Status:      profile.Status,
		LastLoginAt: profile.LastLoginAt,
		CreatedAt:   profile.CreatedAt,
	}

	return response.JSON(c, fiber.StatusOK, "Current admin profile retrieved", resp)
}

// Logout terminates the current administrative session.
// It revokes the session in Redis (preventing reuse) and instructs the client browser
// to clear the admin_session cookie with MaxAge = -1.
//
// Responses:
//   - 200 OK: Successfully logged out and cookie cleared.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sessionID := c.Cookies(AdminSessionCookie)
	if sessionID == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			sessionID = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if sessionID != "" {
		_ = h.authUseCase.Logout(c.Context(), sessionID)
	}

	cookieDomain := h.baseConfig.CookieDomain
	if cookieDomain == "localhost" || cookieDomain == "127.0.0.1" {
		cookieDomain = ""
	}

	c.Cookie(&fiber.Cookie{
		Name:     AdminSessionCookie,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
		HTTPOnly: true,
		Secure:   h.baseConfig.APP_ENV == "production" || h.baseConfig.APP_ENV == "staging",
	})

	return response.JSON(c, fiber.StatusOK, "Logged out successfully", nil)
}
