package routes

import (
	"sage-backend/internal/shared/response"
	"sage-backend/internal/users/handlers"
	"sage-backend/internal/users/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App, 
	ah *handlers.AuthHandler, 
	ch *handlers.CompanyHandler,
	m *middlewares.AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE API SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	RegisterAuthRoutes(v1, ah, m)
	RegisterCompanyRoutes(v1, ch, m)
	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}

func RegisterAuthRoutes(router fiber.Router, ah *handlers.AuthHandler, m *middlewares.AuthMiddleware) {
	auth := router.Group("/auth")

	auth.Post("/register", ah.CreateUser)
	auth.Post("/login", ah.Login)
	auth.Get("/login/:provider", ah.BeginAuthLogin)
	auth.Get("/callback/:provider", ah.AuthCallback)

	auth.Get("/generate-2fa", m.RequireAuth, ah.Generate2FA)
	auth.Post("/enable-2fa", m.RequireAuth, ah.Enable2FA)
	auth.Post("/verify-2fa", m.RequireAuth, ah.Verify2FA)

	auth.Post("/forgot-password", ah.ForgotPassword)
	auth.Post("/verify-reset-token", ah.VerifyResetToken)
	auth.Post("/reset-password", ah.ResetPassword)

	auth.Post("/send-verification-email", ah.SendVerificationEmail)
	auth.Post("/verify-email", ah.VerifyEmail)
}

func RegisterCompanyRoutes(router fiber.Router, ch *handlers.CompanyHandler, m *middlewares.AuthMiddleware) {
	company := router.Group("/company")

	company.Get("/industries", ch.GetIndustries)
	company.Get("/organization-roles", ch.GetOrganizationRoles)

	company.Post("/invite", m.RequireAuth, ch.InviteMember)
	company.Get("/invitations/accept", m.RequireAuth, ch.AcceptInvitation)
}
