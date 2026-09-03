package usecase

import (
	"context"
	"log"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase/dto"
	"time"

	"github.com/redis/go-redis/v9"
)

type PasswordReset struct {
	userRepo outbound.UserRepository
	redis    *redis.Client
}

func NewPasswordReset(userRepo outbound.UserRepository, redis *redis.Client) *PasswordReset {
	return &PasswordReset{
		userRepo: userRepo,
		redis:    redis,
	}
}

func (p *PasswordReset) ForgotPassword(ctx context.Context, email string) error {
	em, err := domain.NewEmail(email)
	if err != nil {
		return err
	}

	token := utils.GenerateSecureOTP()
	err = p.redis.Set(ctx, "password_reset:"+token, em.String(), 15*60*time.Second).Err()
	if err != nil {
		return err
	}
	log.Println("Password reset token for ", em.String(), " : ", token)
	return nil
}

func (p *PasswordReset) VerifyResetToken(ctx context.Context, token string) error {
	_, err := p.redis.Get(ctx, "password_reset:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return apperrors.BadException("invalid or expired token")
		}
		return err
	}
	return nil
}

func (p *PasswordReset) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest, token string) error {
	if req.Password != req.ConfirmPassword {
		return apperrors.BadException("passwords do not match")
	}

	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return err
	}

	emailStr, err := p.redis.Get(ctx, "password_reset:"+token).Result()
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

	_, err = p.userRepo.GetUserByEmail(ctx, email.String())
	if err != nil {
		return err
	}

	rawHash, err := utils.HashPassword(password.String())
	if err != nil {
		return err
	}

	hash := domain.NewPasswordHash(rawHash)

	err = p.userRepo.UpdateUserPassword(ctx, email.String(), hash.String())
	if err != nil {
		return err
	}

	p.redis.Del(ctx, "password_reset:"+token)
	return nil
}
