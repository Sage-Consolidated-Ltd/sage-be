package domain

import "errors"

var (
	ErrAdminNotFound           = errors.New("admin not found")
	ErrAdminAlreadyExists      = errors.New("admin with this email already exists")
	ErrAccountSuspended        = errors.New("admin account is suspended")
	ErrAccountTemporarilyLocked = errors.New("admin account is temporarily locked due to repeated failed login attempts")
	ErrAccountPermanentlyLocked = errors.New("admin account is permanently locked; manual administrator review required")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrInvalidOTP              = errors.New("invalid or expired OTP code")
	ErrOTPExpired              = errors.New("OTP code has expired")
	ErrIPMismatch              = errors.New("session IP mismatch detected; security policy requires re-authentication")
	ErrSessionExpired          = errors.New("admin session expired or revoked")
	ErrEmptyEmail              = errors.New("email cannot be empty")
	ErrInvalidEmail            = errors.New("invalid email format: must match a valid email pattern")
)
