package routes

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/users/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App,
	ah *handlers.AuthHandler,
	ch *handlers.CompanyHandler,
	ph *handlers.ProfileHandler,
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
	RegisterProfileRoutes(v1, ph, m)
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
	// Legacy /company routes
	company := router.Group("/company")
	company.Get("/industries", ch.GetIndustries)
	company.Get("/organization-roles", ch.GetOrganizationRoles)
	company.Post("/invite", m.RequireAuth, ch.InviteMember)
	company.Get("/invitations/accept", m.RequireAuth, ch.AcceptInvitation)

	// Organization routes per AGENTS.md spec
	org := router.Group("/organization", m.RequireAuth)
	org.Get("/", ch.GetOrganization)
	org.Patch("/", ch.UpdateOrganization)

	// Members
	members := org.Group("/members")
	members.Get("/", ch.ListMembers)
	members.Post("/invite", ch.InviteMembers)
	members.Patch("/:id/role", ch.UpdateMemberRole)
	members.Delete("/:id", ch.RemoveMember)

	// Settings
	settings := org.Group("/settings")
	settings.Get("/", ch.GetOrganizationSettings)
	settings.Patch("/", ch.UpdateOrganizationSettings)

	// Permissions & Custom Roles
	org.Get("/permissions", ch.ListPermissions)
	org.Get("/permission-groups", ch.ListPermissionGroups)

	customRoles := org.Group("/custom-roles")
	customRoles.Get("/", ch.ListCustomRoles)
	customRoles.Post("/", ch.CreateCustomRole)
	customRoles.Get("/:id", ch.GetCustomRole)
	customRoles.Patch("/:id", ch.UpdateCustomRole)
	customRoles.Delete("/:id", ch.DeleteCustomRole)
}

func RegisterProfileRoutes(router fiber.Router, ph *handlers.ProfileHandler, m *middlewares.AuthMiddleware) {
	profile := router.Group("/profile", m.RequireAuth)

	// Identity - /api/v1/profile
	profile.Get("/", ph.GetProfile)
	profile.Patch("/", ph.UpdateProfile)

	// Legacy compatibility - /api/v1/profile/me
	profile.Get("/me", ph.GetProfile)
	profile.Patch("/me", ph.UpdateProfile)

	profile.Post(
		"/avatar",
		middlewares.ValidateFileUpload(
			"avatar",
			middlewares.MaxAvatarSize,
			middlewares.AvatarMimeTypes,
		),
		ph.UploadAvatar,
	)

	// Preferences - /api/v1/profile/preferences
	profile.Get("/preferences", ph.GetPreferences)
	profile.Patch("/preferences", ph.UpdatePreferences)

	// Notifications - /api/v1/profile/notifications
	profile.Get("/notifications", ph.GetNotifications)
	profile.Patch("/notifications", ph.UpdateNotifications)

	// Sessions - /api/v1/profile/sessions
	profile.Get("/sessions", ph.GetSessions)
	profile.Delete("/sessions/:id", ph.RevokeSession)

	// Activity - /api/v1/profile/activity
	profile.Get("/activity", ph.GetActivity)
}
