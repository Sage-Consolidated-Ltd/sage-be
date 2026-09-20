package email

import (
	"context"

	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/shared/mailer"
)

const adminOTPTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Sage Admin Verification</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0f172a; padding: 30px; color: #f8fafc;">
  <div style="max-width: 480px; margin: 0 auto; background-color: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 32px;">
    <h2 style="margin-top: 0; color: #38bdf8; font-size: 20px;">Sage Platform Admin Access</h2>
    <p style="color: #cbd5e1; font-size: 15px; line-height: 1.5;">Use the one-time verification code below to complete your administrator sign in:</p>
    <div style="background-color: #0f172a; border: 1px solid #475569; border-radius: 8px; padding: 18px; text-align: center; margin: 24px 0;">
      <span style="font-family: monospace; font-size: 32px; font-weight: 700; letter-spacing: 8px; color: #38bdf8;">{{.OTP}}</span>
    </div>
    <p style="color: #94a3b8; font-size: 13px; margin-bottom: 0;">This code expires in 10 minutes. If you did not initiate this request, an unauthorized party may know your credentials. Please take immediate action.</p>
  </div>
</body>
</html>`

type AdminEmailData struct {
	OTP string
}

type EmailServiceAdapter struct {
	client mailer.EmailClientInt
}

func NewEmailServiceAdapter(client mailer.EmailClientInt) outbound.EmailService {
	return &EmailServiceAdapter{client: client}
}

var _ outbound.EmailService = (*EmailServiceAdapter)(nil)

func (s *EmailServiceAdapter) SendAdminOTP(ctx context.Context, toEmail string, otpCode string) error {
	data := AdminEmailData{OTP: otpCode}
	return s.client.SendEmail(
		[]string{toEmail},
		"Sage Admin Authentication - Verification Code",
		"admin_otp",
		adminOTPTemplate,
		data,
	)
}
