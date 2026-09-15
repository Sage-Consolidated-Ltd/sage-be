package outbound

import (
	"context"
	"time"
)

type OTPPendingData struct {
	AdminID    string
	OTP        string
	RememberMe bool
	ClientIP   string
}

type OTPStore interface {
	SaveOTP(ctx context.Context, data OTPPendingData, ttl time.Duration) error
	VerifyAndConsumeOTP(ctx context.Context, adminID string, otp string) (*OTPPendingData, error)
	ClearOTP(ctx context.Context, adminID string) error
}
