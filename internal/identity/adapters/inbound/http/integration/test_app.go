package integration

import (
	"testing"

	"sage-backend/internal/identity"
	identity_http "sage-backend/internal/identity/adapters/inbound/http"
	"sage-backend/internal/organization"
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

func setUpUsersApp(t *testing.T) *TestHarness {
	t.Helper()
	cfg := config.SetupTestConfig()

	// Infrastructure
	database := testPostgres(t)
	rdb := testRedis(t)

	// Session store
	config.InitSessionStore(&cfg.BaseConfig)

	// Services
	jwtService := &jwt.JwtService{JwtSecret: cfg.JWTSecret}
	noOpEmail := &NoOpEmailClient{}

	// Modules
	orgMod := organization.NewModule(database, nil, noOpEmail, rdb, &cfg.BaseConfig, nil)
	identityMod := identity.NewModule(database, rdb, cfg, jwtService, noOpEmail, nil, orgMod.CompanyRepo)

	// Middleware
	middleware := &middlewares.AuthMiddleware{}

	// Fiber app
	app := fiber.New()

	// Routes
	setUpRoutes(app, identityMod.AuthHandler, identityMod.ProfileHandler, middleware)

	return &TestHarness{
		App:      app,
		Database: database,
		Redis:    rdb,
	}
}

func setUpRoutes(
	app *fiber.App,
	ah *identity_http.AuthHandler,
	ph *identity_http.ProfileHandler,
	m *middlewares.AuthMiddleware,
) {
	testGroup := app.Group("/test/v1")
	identity_http.SetUpRouter(testGroup, ah, ph, m)
}
