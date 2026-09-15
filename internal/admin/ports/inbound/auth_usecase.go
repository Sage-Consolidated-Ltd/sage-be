package inbound

import (
	"context"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/admin/usecase/dto"
)

type AuthUseCase interface {
	Login(ctx context.Context, input dto.AdminLoginInput) (*dto.AdminLoginResult, error)
	VerifyOTP(ctx context.Context, input dto.VerifyOTPInput) (*dto.VerifyOTPResult, error)
	ResendOTP(ctx context.Context, input dto.ResendOTPInput) error
	ValidateSession(ctx context.Context, sessionID string, clientIP string) (*outbound.AdminSession, error)
	GetAdminProfile(ctx context.Context, adminID string) (*dto.AdminProfileResult, error)
	Logout(ctx context.Context, sessionID string) error
	BootstrapSuperAdmin(ctx context.Context, email, password string) error
}
