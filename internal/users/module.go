package users

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/storage/s3"
	users_http "sage-backend/internal/users/adapters/inbound/http"
	"sage-backend/internal/users/adapters/outbound/postgres"
	"sage-backend/internal/users/ports/inbound"
	"sage-backend/internal/users/usecase"
	"sage-backend/pkg/jwt"

	"github.com/redis/go-redis/v9"
)

type Module struct {
	AuthUseCase    inbound.AuthUseCase
	UserUseCase    inbound.UserUseCase
	CompanyUseCase inbound.CompanyUseCase

	AuthHandler    *users_http.AuthHandler
	CompanyHandler *users_http.CompanyHandler
	ProfileHandler *users_http.ProfileHandler
}

func NewModule(
	database *db.DB,
	redisClient *redis.Client,
	appConfig *config.APIConfig,
	jwtService *jwt.JwtService,
	emailClient mailer.EmailClientInt,
	uploader *s3.Uploader,
) *Module {
	userRepo := postgres.NewUsersRepository(database)
	profileRepo := postgres.NewProfileRepository(database)
	companyRepo := postgres.NewCompanyRepository(database)

	authUseCase := usecase.NewAuthService(
		userRepo,
		jwtService,
		appConfig,
		redisClient,
		companyRepo,
		emailClient,
	)

	userUseCase := usecase.NewUsersServices(
		userRepo,
		profileRepo,
		redisClient,
		uploader,
	)

	companyUseCase := usecase.NewCompanyServices(
		companyRepo,
		userRepo,
		emailClient,
		redisClient,
		&appConfig.BaseConfig,
	)

	oAuthConfig := config.NewOAuthConfig(
		appConfig.GoogleClientId, appConfig.GoogleClientSecret, appConfig.GoogleRedirectUrl,
		appConfig.GitHubClientId, appConfig.GitHubClientSecret, appConfig.GitHubRedirectUrl,
		appConfig.AzureClientId, appConfig.AzureClientSecret, appConfig.AzureRedirectUrl,
	)

	authHandler := users_http.NewAuthHandler(authUseCase, oAuthConfig, appConfig)
	companyHandler := users_http.NewCompanyHandler(companyUseCase)
	profileHandler := users_http.NewProfileHandler(userUseCase)

	return &Module{
		AuthUseCase:    authUseCase,
		UserUseCase:    userUseCase,
		CompanyUseCase: companyUseCase,
		AuthHandler:    authHandler,
		CompanyHandler: companyHandler,
		ProfileHandler: profileHandler,
	}
}
