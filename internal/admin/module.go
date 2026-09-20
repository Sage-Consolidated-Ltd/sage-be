package admin

import (
	adminHttp "sage-backend/internal/admin/adapters/inbound/http"
	"sage-backend/internal/admin/adapters/outbound/email"
	"sage-backend/internal/admin/adapters/outbound/postgres"
	adminRedis "sage-backend/internal/admin/adapters/outbound/redis"
	"sage-backend/internal/admin/ports/inbound"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/admin/usecase"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/mailer"

	"github.com/redis/go-redis/v9"
)

type Module struct {
	AdminRepo      outbound.AdminRepository
	OTPStore       outbound.OTPStore
	SessionStore   outbound.SessionStore
	AuthUseCase    inbound.AuthUseCase
	AuthHandler    *adminHttp.AuthHandler
	AuthMiddleware *adminHttp.AdminAuthMiddleware
}

func NewModule(
	database *db.DB,
	redisClient *redis.Client,
	emailClient mailer.EmailClientInt,
	baseConfig *config.BaseConfig,
) *Module {
	adminRepo := postgres.NewAdminRepository(database)
	otpStore := adminRedis.NewRedisOTPStore(redisClient)
	sessionStore := adminRedis.NewRedisSessionStore(redisClient)
	emailService := email.NewEmailServiceAdapter(emailClient)

	authUseCase := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
	authHandler := adminHttp.NewAuthHandler(authUseCase, baseConfig)
	authMiddleware := adminHttp.NewAdminAuthMiddleware(authUseCase)

	return &Module{
		AdminRepo:      adminRepo,
		OTPStore:       otpStore,
		SessionStore:   sessionStore,
		AuthUseCase:    authUseCase,
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
	}
}