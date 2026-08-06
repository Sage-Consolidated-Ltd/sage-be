package usecase

import (
	"context"
	"log"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/outbound"
	"time"

	"github.com/redis/go-redis/v9"
)

type EmailVerification struct {
	userRepo outbound.UserRepository
	redis    *redis.Client
	mailer   mailer.EmailClientInt
}

func NewEmailVerification(userRepo outbound.UserRepository, redis *redis.Client, mailer mailer.EmailClientInt) *EmailVerification {
	return &EmailVerification{
		userRepo: userRepo,
		redis:    redis,
		mailer:   mailer,
	}
}

func (e *EmailVerification) SendEmailVerification(ctx context.Context, emailStr string) error {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return err
	}

	user, err := e.userRepo.GetUserByEmail(ctx, email.String())
	if err != nil {
		return err
	}
	if user.IsVerified() {
		return apperrors.BadException("email is already verified")
	}

	token := utils.GenerateSecureOTP()
	err = e.redis.Set(ctx, "email_verification:"+token, email.String(), 24*time.Hour).Err()
	if err != nil {
		return err
	}

	if err := e.mailer.SendVerificationEmail([]string{email.String()}, mailer.VerificationEmailData{
		Name:      user.FirstName() + " " + user.LastName(),
		OTP:       token,
		ExpiresIn: "24 hours",
	}); err != nil {
		return err
	}
	log.Println("Email verification token for ", email.String(), " : ", token)
	return nil
}

func (e *EmailVerification) VerifyEmail(ctx context.Context, token string) error {
	emailStr, err := e.redis.Get(ctx, "email_verification:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return apperrors.BadException("invalid or expired token")
		}
		return err
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return err
	}

	err = e.userRepo.MarkEmailVerified(ctx, email.String())
	if err != nil {
		return err
	}

	e.redis.Del(ctx, "email_verification:"+token)
	return nil
}
