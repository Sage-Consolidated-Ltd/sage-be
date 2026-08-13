package integration

import (
	"testing"

	"sage-backend/internal/identity"
	identity_http "sage-backend/internal/identity/adapters/inbound/http"
	"sage-backend/internal/organization"
	org_http "sage-backend/internal/organization/adapters/inbound/http"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	redisClient "github.com/redis/go-redis/v9"
)

type TestHarness struct {
	App      *fiber.App
	Database *db.DB
	Redis    *redisClient.Client
}

type NoOpEmailClient struct{}

func (n *NoOpEmailClient) SendEmail(to []string, subject, templateName, rawTemplate string, data any) error {
	return nil
}

func (n *NoOpEmailClient) SendMemberInvitationEmail(to []string, data mailer.MemberInvitationEmailData) error {
	return nil
}

func (n *NoOpEmailClient) SendVerificationEmail(to []string, data mailer.VerificationEmailData) error {
	return nil
}

func setUpOrgApp(t *testing.T) *TestHarness {
	t.Helper()
	cfg := config.SetupTestConfig()

	database := testPostgres(t)
	rdb := testRedis(t)
	if database == nil || rdb == nil {
		t.Skip("Integration test skipped: database or Redis instance unavailable")
		return nil
	}

	config.InitSessionStore(&cfg.BaseConfig)

	jwtService := &jwt.JwtService{JwtSecret: cfg.JWTSecret}
	noOpEmail := &NoOpEmailClient{}

	orgMod := organization.NewModule(database, nil, noOpEmail, rdb, &cfg.BaseConfig, nil)
	identityMod := identity.NewModule(database, rdb, cfg, jwtService, noOpEmail, nil, orgMod.CompanyRepo)

	middleware := &middlewares.AuthMiddleware{}

	app := fiber.New()

	testGroup := app.Group("/test/v1")
	identity_http.SetUpRouter(testGroup, identityMod.AuthHandler, identityMod.ProfileHandler, middleware)
	org_http.SetUpRouter(testGroup, orgMod.CompanyHandler, middleware)

	return &TestHarness{
		App:      app,
		Database: database,
		Redis:    rdb,
	}
}
