package http

import (
	_ "sage-backend/internal/admin/adapters/inbound/http/dto"
	_ "sage-backend/internal/shared/response"
)

// @Summary      Admin Login Challenge
// @Description  Initiates an administrative authentication challenge with email and password. Upon successful credential verification, a 6-digit one-time password (OTP) is dispatched to the administrator's registered email address.
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AdminLoginRequest  true  "Admin Login Credentials"
// @Success      200      {object}  response.Response{data=dto.AdminLoginResponse}  "OTP challenge initiated successfully"
// @Failure      400      {object}  response.Response  "Invalid request body or JSON syntax"
// @Failure      401      {object}  response.Response  "Invalid email or password"
// @Failure      403      {object}  response.Response  "Account is temporarily locked (after 5 failed attempts), permanently locked (after 10 failed attempts), or suspended"
// @Failure      422      {object}  response.Response  "Validation error (e.g. invalid email format or password length < 8)"
// @Failure      500      {object}  response.Response  "Internal server error"
// @Router       /admin/auth/login [post]
func _AdminLogin() {}

// @Summary      Verify Admin OTP
// @Description  Verifies the 6-digit OTP code dispatched during the login challenge. On success, sets a secure, HTTP-only session cookie (admin_session) and returns the authenticated administrator profile and session metadata.
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyOTPRequest  true  "OTP Verification Payload"
// @Success      200      {object}  response.Response{data=dto.AdminAuthResponse}  "Authentication successful, session established"
// @Failure      400      {object}  response.Response  "Invalid or expired OTP verification code"
// @Failure      403      {object}  response.Response  "Account locked"
// @Failure      422      {object}  response.Response  "Validation error (e.g. admin_id not a UUID or OTP not 6 digits)"
// @Failure      500      {object}  response.Response  "Internal server error"
// @Router       /admin/auth/verify-otp [post]
func _AdminVerifyOTP() {}

// @Summary      Resend Admin OTP
// @Description  Dispatches a new 6-digit OTP verification code to the administrator's email address if the previous code expired or was not received.
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ResendOTPRequest  true  "Resend OTP Payload"
// @Success      200      {object}  response.Response  "New verification code dispatched"
// @Failure      400      {object}  response.Response  "Invalid request body"
// @Failure      403      {object}  response.Response  "Account locked"
// @Failure      404      {object}  response.Response  "Admin record not found"
// @Failure      422      {object}  response.Response  "Validation error (e.g. admin_id not a UUID)"
// @Failure      500      {object}  response.Response  "Internal server error"
// @Router       /admin/auth/resend-otp [post]
func _AdminResendOTP() {}

// @Summary      Get Current Admin Profile
// @Description  Retrieves profile information for the currently authenticated administrator identified via the admin_session cookie or Authorization Bearer header.
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Security     AdminSessionAuth
// @Security     AdminBearerAuth
// @Success      200      {object}  response.Response{data=dto.AdminProfileResponse}  "Current admin profile retrieved"
// @Failure      401      {object}  response.Response  "Unauthorized: Missing, invalid, expired, or IP-mismatched admin session"
// @Failure      404      {object}  response.Response  "Admin profile not found"
// @Router       /admin/auth/me [get]
func _AdminMe() {}

// @Summary      Admin Logout
// @Description  Invalidates the active administrator session in the Redis session store and destroys the admin_session HTTP-only cookie.
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Security     AdminSessionAuth
// @Security     AdminBearerAuth
// @Success      200      {object}  response.Response  "Logged out successfully"
// @Router       /admin/auth/logout [post]
func _AdminLogout() {}

