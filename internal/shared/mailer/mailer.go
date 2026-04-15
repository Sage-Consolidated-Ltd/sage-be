package mailer

import (
	"fmt"
	"sage-backend/internal/shared/config"

	"github.com/resendlabs/resend-go"
)

type EmailClient struct {
	client *resend.Client
	from string
}
type EmailClientInt interface {
	SendEmail(to []string, subject, templateName, rawTemplate string, data any) error
	SendMemberInvitationEmail(to []string, data MemberInvitationEmailData) error
	SendVerificationEmail(to []string, data VerificationEmailData) error
}

func NewEmailClient(cfg *config.BaseConfig) EmailClientInt {
	return &EmailClient{
		client: resend.NewClient(cfg.ResendApiKey),
		from:   cfg.ResendFromEmail,
	}
}

func (e *EmailClient) SendEmail(to []string, subject, templateName, rawTemplate string, data any) error {
	html, err := RenderTemplate(templateName, rawTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

    _, err = e.client.Emails.Send(&resend.SendEmailRequest{
        From:    e.from,
        To:      to,
        Subject: subject,
        Html:    html,
    })
    if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}
func (e *EmailClient) SendMemberInvitationEmail(to []string, data MemberInvitationEmailData) error {
	subject := fmt.Sprintf("You're invited to join %s on Sage!", data.OrganizationName)
	return e.SendEmail(to, subject, "member_invitation", MEMBER_INVITE_TEMPLATE, data)
}
func (e *EmailClient) SendVerificationEmail(to []string, data VerificationEmailData) error {
	subject := "Verify your email address"
	return e.SendEmail(to, subject, "verification", VERIFICATION_EMAIL_TEMPLATE, data)
}