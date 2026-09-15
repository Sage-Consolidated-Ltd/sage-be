package dto

import "time"

type AdminLoginInput struct {
	Email      string
	Password   string
	RememberMe bool
	ClientIP   string
}

type AdminLoginResult struct {
	AdminID string
	Message string
}

type VerifyOTPInput struct {
	AdminID  string
	OTP      string
	ClientIP string
}

type VerifyOTPResult struct {
	SessionID string
	ExpiresAt time.Time
	Admin     AdminProfileResult
}

type ResendOTPInput struct {
	AdminID  string
	ClientIP string
}

type AdminProfileResult struct {
	ID          string
	Email       string
	Role        string
	Status      string
	LastLoginAt *time.Time
	CreatedAt   time.Time
}
