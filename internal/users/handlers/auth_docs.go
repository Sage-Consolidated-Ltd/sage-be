package handlers

import (
	_ "sage-backend/internal/users/models"
	_ "sage-backend/internal/users/requests"
)

type Generate2FAResponse struct {
	QRCode string `json:"qr_code" example:"data:image/png;base64,..."`
	Secret string `json:"secret" example:"JBSWY3DPEHPK3PXP....."`
}

// @Summary Create User
// @Description Creates a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.OnboardingRequest true "User Details"
// @Success 201
// @Router /auth/register [post]
func _CreateUser() {}

// @Summary      Login User with OAUTH
// @Description  Logs in a user using OAUTH (Google, Github)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        provider  path      string  true  "OAuth Provider (e.g. google, github)"
// @Success      200
// @Router       /auth/login/{provider} [get]
func _BeginAuthLogin() {}

// @Summary      Login User
// @Description  Logs in a user using email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.LoginRequest true "Login Credentials"
// @Success      200       {object}  models.GetUserResponse
// @Router       /auth/login [post]
func _Login() {}

// @Summary      Generate 2FA Secret
// @Description  Generates a 2FA secret for the user and returns the otpauth URL and the base32 secret
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security 	SessionAuth
// @Success      200     {object}  Generate2FAResponse
// @Router       /auth/generate-2fa [get]
func _Generate2FA() {}

// @Summary      Enable 2FA
// @Description  Enables 2FA for the user after verifying the provided TOTP code
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.GoogleAuthenticatorRequest true "TOTP Code"
// @Security 	SessionAuth
// @Success      200
// @Router       /auth/enable-2fa [post]
func _Enable2FA() {}

// @Summary      Verify 2FA
// @Description  Verifies the provided TOTP code for the user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.GoogleAuthenticatorRequest true "TOTP Code"
// @Security 	SessionAuth
// @Success      200
// @Router       /auth/verify-2fa [post]
func _Verify2FA() {}

// @Summary 	Forgot Password
// @Description  Initiates the password reset process for a user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.ForgotPasswordRequest true "Email Address"
// @Success      200
// @Router       /auth/forgot-password [post]
func _ForgotPassword() {}

// @Summary 	Verify Reset Token
// @Description  Verifies the password reset token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.VerifyResetTokenRequest true "Reset Token"
// @Success      200
// @Router       /auth/verify-reset-token [post]
func _VerifyResetToken() {}

// @Summary 	Reset Password
// @Description  Resets the user's password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.ResetPasswordRequest true "New Password"
// @Success      200
// @Router       /auth/reset-password [post]
func _ResetPassword() {}

// @Summary 	Send Verification Email
// @Description  Sends an email with a verification code to the user's email address
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.SendVerificationEmailRequest true "Email Address"
// @Success      200
// @Router       /auth/send-verification-email [post]
func _SendVerificationEmail() {}

// @Summary 	Verify Email
// @Description  Verifies the user's email address
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      requests.VerifyEmailRequest true "Verification Token"
// @Success      200
// @Router       /auth/verify-email [post]
func _VerifyEmail() {}
