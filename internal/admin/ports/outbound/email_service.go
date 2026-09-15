package outbound

import "context"

type EmailService interface {
	SendAdminOTP(ctx context.Context, toEmail string, otpCode string) error
}
