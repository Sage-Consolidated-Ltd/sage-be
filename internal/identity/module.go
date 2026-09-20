package identity

import (
	"sage-backend/internal/identity/adapters/inbound/http"
	"sage-backend/internal/identity/adapters/outbound/postgres"
	"sage-backend/internal/identity/ports/inbound"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase"
	orgOutbound "sage-backend/internal/organization/ports/outbound"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/pkg/jwt"

	"github.com/redis/go-redis/v9"
)

type Module struct {
	UserRepo       outbound.UserRepository
	ProfileRepo    outbound.ProfileRepository
	AuthUseCase    inbound.AuthUseCase
	UserUseCase    inbound.UserUseCase
	AuthHandler    *http.AuthHandler
	ProfileHandler *http.ProfileHandler
}

func NewModule(
	database *db.DB,
	redisClient *redis.Client,
	appConfig *config.APIConfig,
	jwtService *jwt.JwtService,
	emailClient mailer.EmailClientInt,
	uploader *s3.Uploader,
	companyRepo orgOutbound.CompanyRepository,
) *Module {
	userRepo := postgres.NewUserRepository(database)
	profileRepo := postgres.NewProfileRepository(database)

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

	authHandler := http.NewAuthHandler(authUseCase, appConfig)
	profileHandler := http.NewProfileHandler(userUseCase)

	return &Module{
		UserRepo:       userRepo,
		ProfileRepo:    profileRepo,
		AuthUseCase:    authUseCase,
		UserUseCase:    userUseCase,
		AuthHandler:    authHandler,
		ProfileHandler: profileHandler,
	}
}
