package organization

import (
	identityOutbound "sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/organization/adapters/inbound/http"
	"sage-backend/internal/organization/adapters/outbound/postgres"
	"sage-backend/internal/organization/ports/inbound"
	"sage-backend/internal/organization/ports/outbound"
	"sage-backend/internal/organization/usecase"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/storage/s3"

	"github.com/redis/go-redis/v9"
)

type Module struct {
	CompanyRepo    outbound.CompanyRepository
	CompanyUseCase inbound.CompanyUseCase
	CompanyHandler *http.CompanyHandler
}

func NewModule(
	database *db.DB,
	userRepo identityOutbound.UserRepository,
	mailer mailer.EmailClientInt,
	redisClient *redis.Client,
	config *config.BaseConfig,
	uploader *s3.Uploader,
) *Module {
	companyRepo := postgres.NewCompanyRepository(database)
	companyUseCase := usecase.NewCompanyServices(companyRepo, userRepo, mailer, redisClient, config, uploader)
	companyHandler := http.NewCompanyHandler(companyUseCase)

	return &Module{
		CompanyRepo:    companyRepo,
		CompanyUseCase: companyUseCase,
		CompanyHandler: companyHandler,
	}
}
