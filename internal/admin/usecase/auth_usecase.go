package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/inbound"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/admin/usecase/dto"
	"sage-backend/internal/shared/utils"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	adminRepo    outbound.AdminRepository
	otpStore     outbound.OTPStore
	sessionStore outbound.SessionStore
	emailService outbound.EmailService
}

func NewAuthUseCase(
	adminRepo outbound.AdminRepository,
	otpStore outbound.OTPStore,
	sessionStore outbound.SessionStore,
	emailService outbound.EmailService,
) inbound.AuthUseCase {
	return &AuthUseCase{
		adminRepo:    adminRepo,
		otpStore:     otpStore,
		sessionStore: sessionStore,
		emailService: emailService,
	}
}

var _ inbound.AuthUseCase = (*AuthUseCase)(nil)

func (u *AuthUseCase) Login(ctx context.Context, input dto.AdminLoginInput) (*dto.AdminLoginResult, error) {
	now := time.Now().UTC()
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}

	admin, err := u.adminRepo.GetByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := admin.CanLogin(now); err != nil {
		return nil, err
	}

	if !utils.CompareHashAndPassword(input.Password, admin.PasswordHash()) {
		locked, lockStatus := admin.RecordFailedAttempt(now)
		_ = u.adminRepo.Update(ctx, admin)

		if locked {
			if lockStatus == domain.StatusPermanentlyLocked {
				return nil, domain.ErrAccountPermanentlyLocked
			}
			return nil, domain.ErrAccountTemporarilyLocked
		}
		return nil, domain.ErrInvalidCredentials
	}

	otp := utils.GenerateSecureOTP()
	otpData := outbound.OTPPendingData{
		AdminID:    admin.ID(),
		OTP:        otp,
		RememberMe: input.RememberMe,
		ClientIP:   input.ClientIP,
	}

	if err := u.otpStore.SaveOTP(ctx, otpData, 10*time.Minute); err != nil {
		return nil, err
	}

	if err := u.emailService.SendAdminOTP(ctx, admin.Email().String(), otp); err != nil {
		return nil, err
	}

	return &dto.AdminLoginResult{
		AdminID: admin.ID(),
		Message: "Verification code sent to your email",
	}, nil
}

func (u *AuthUseCase) VerifyOTP(ctx context.Context, input dto.VerifyOTPInput) (*dto.VerifyOTPResult, error) {
	now := time.Now().UTC()

	admin, err := u.adminRepo.GetByID(ctx, input.AdminID)
	if err != nil {
		return nil, err
	}

	if err := admin.CanLogin(now); err != nil {
		return nil, err
	}

	pending, err := u.otpStore.VerifyAndConsumeOTP(ctx, input.AdminID, input.OTP)
	if err != nil {
		return nil, err
	}

	admin.RecordSuccessfulLogin(now)
	if err := u.adminRepo.Update(ctx, admin); err != nil {
		return nil, err
	}

	ttl := 12 * time.Hour
	if pending.RememberMe {
		ttl = 14 * 24 * time.Hour
	}

	sessionID := uuid.New().String()
	clientIP := input.ClientIP
	if clientIP == "" {
		clientIP = pending.ClientIP
	}

	session := &outbound.AdminSession{
		SessionID: sessionID,
		AdminID:   admin.ID(),
		Role:      admin.Role().String(),
		Email:     admin.Email().String(),
		BoundIP:   clientIP,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	if err := u.sessionStore.CreateSession(ctx, session, ttl); err != nil {
		return nil, err
	}

	return &dto.VerifyOTPResult{
		SessionID: sessionID,
		ExpiresAt: session.ExpiresAt,
		Admin: dto.AdminProfileResult{
			ID:          admin.ID(),
			Email:       admin.Email().String(),
			Role:        admin.Role().String(),
			Status:      admin.Status().String(),
			LastLoginAt: admin.LastLoginAt(),
			CreatedAt:   admin.CreatedAt(),
		},
	}, nil
}

func (u *AuthUseCase) ResendOTP(ctx context.Context, input dto.ResendOTPInput) error {
	now := time.Now().UTC()

	admin, err := u.adminRepo.GetByID(ctx, input.AdminID)
	if err != nil {
		return err
	}

	if err := admin.CanLogin(now); err != nil {
		return err
	}

	otp := utils.GenerateSecureOTP()
	otpData := outbound.OTPPendingData{
		AdminID:  admin.ID(),
		OTP:      otp,
		ClientIP: input.ClientIP,
	}

	if err := u.otpStore.SaveOTP(ctx, otpData, 10*time.Minute); err != nil {
		return err
	}

	return u.emailService.SendAdminOTP(ctx, admin.Email().String(), otp)
}

func (u *AuthUseCase) ValidateSession(ctx context.Context, sessionID string, clientIP string) (*outbound.AdminSession, error) {
	session, err := u.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Anomaly detection: if session was bound to an IP and the incoming IP doesn't match, revoke immediately!
	if session.BoundIP != "" && clientIP != "" && session.BoundIP != clientIP {
		_ = u.sessionStore.DeleteSession(ctx, sessionID)
		return nil, domain.ErrIPMismatch
	}

	return session, nil
}

func (u *AuthUseCase) GetAdminProfile(ctx context.Context, adminID string) (*dto.AdminProfileResult, error) {
	admin, err := u.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	return &dto.AdminProfileResult{
		ID:          admin.ID(),
		Email:       admin.Email().String(),
		Role:        admin.Role().String(),
		Status:      admin.Status().String(),
		LastLoginAt: admin.LastLoginAt(),
		CreatedAt:   admin.CreatedAt(),
	}, nil
}

func (u *AuthUseCase) Logout(ctx context.Context, sessionID string) error {
	return u.sessionStore.DeleteSession(ctx, sessionID)
}

// BootstrapSuperAdmin idempotently ensures the initial Super Admin account exists.
// If an admin with the provided email already exists, it safely no-ops.
// Otherwise, it validates credentials, hashes the password with bcrypt, and creates
// the Super Admin record in active status.
func (u *AuthUseCase) BootstrapSuperAdmin(ctx context.Context, email, password string) error {
	now := time.Now().UTC()
	emailVO, err := domain.NewEmail(email)
	if err != nil {
		return fmt.Errorf("invalid super admin email: %w", err)
	}

	if len(strings.TrimSpace(password)) < 8 {
		return errors.New("super admin password must be at least 8 characters")
	}

	existing, err := u.adminRepo.GetByEmail(ctx, emailVO.String())
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && !errors.Is(err, domain.ErrAdminNotFound) {
		return fmt.Errorf("failed to check existing super admin: %w", err)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash super admin password: %w", err)
	}

	admin := domain.NewAdmin(
		uuid.NewString(),
		emailVO,
		hash,
		domain.RoleSuperAdmin,
		now,
	)

	if err := u.adminRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("failed to persist initial super admin: %w", err)
	}

	return nil
}

