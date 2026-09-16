package dto

import "time"

type AdminLoginRequest struct {
	Email      string `json:"email" validate:"required,email" example:"admin@sageconsolidated.com"`
	Password   string `json:"password" validate:"required,min=8" example:"SuperSecret123!"`
	RememberMe bool   `json:"remember_me" example:"false"`
}

type AdminLoginResponse struct {
	AdminID string `json:"admin_id" example:"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"`
	Message string `json:"message" example:"Verification OTP has been sent to your registered email"`
}

type VerifyOTPRequest struct {
	AdminID string `json:"admin_id" validate:"required,uuid" example:"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"`
	OTP     string `json:"otp" validate:"required,len=6,numeric" example:"123456"`
}

type ResendOTPRequest struct {
	AdminID string `json:"admin_id" validate:"required,uuid" example:"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"`
}

type AdminProfileResponse struct {
	ID          string     `json:"id" example:"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"`
	Email       string     `json:"email" example:"admin@sageconsolidated.com"`
	Role        string     `json:"role" example:"super_admin"`
	Status      string     `json:"status" example:"active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2026-09-14T12:00:00Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-01-01T00:00:00Z"`
}

type AdminAuthResponse struct {
	Admin     AdminProfileResponse `json:"admin"`
	SessionID string               `json:"session_id" example:"sess_9876543210"`
	ExpiresAt time.Time            `json:"expires_at" example:"2026-09-14T20:00:00Z"`
}
