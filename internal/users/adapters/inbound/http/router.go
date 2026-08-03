package http

import (
	"sage-backend/internal/shared/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetUpRouter(router fiber.Router, ah *AuthHandler, ch *CompanyHandler, ph *ProfileHandler, m *middlewares.AuthMiddleware) {
	RegisterAuthRoutes(router, ah, m)
	RegisterCompanyRoutes(router, ch, m)
	RegisterProfileRoutes(router, ph, m)
}

func RegisterAuthRoutes(router fiber.Router, ah *AuthHandler, m *middlewares.AuthMiddleware) {
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

func RegisterCompanyRoutes(router fiber.Router, ch *CompanyHandler, m *middlewares.AuthMiddleware) {
	company := router.Group("/company")
	company.Get("/industries", ch.GetIndustries)
	company.Get("/organization-roles", ch.GetOrganizationRoles)
	company.Post("/invite", m.RequireAuth, ch.InviteMember)
	company.Get("/invitations/accept", m.RequireAuth, ch.AcceptInvitation)

	org := router.Group("/organization", m.RequireAuth)
	org.Get("/", ch.GetOrganization)
	org.Patch("/", ch.UpdateOrganization)

	members := org.Group("/members")
	members.Get("/", ch.ListMembers)
	members.Post("/invite", ch.InviteMembers)
	members.Patch("/:id/role", ch.UpdateMemberRole)
	members.Delete("/:id", ch.RemoveMember)

	settings := org.Group("/settings")
	settings.Get("/", ch.GetOrganizationSettings)
	settings.Patch("/", ch.UpdateOrganizationSettings)

	org.Get("/permissions", ch.ListPermissions)
	org.Get("/permission-groups", ch.ListPermissionGroups)

	customRoles := org.Group("/custom-roles")
	customRoles.Get("/", ch.ListCustomRoles)
	customRoles.Post("/", ch.CreateCustomRole)
	customRoles.Get("/:id", ch.GetCustomRole)
	customRoles.Patch("/:id", ch.UpdateCustomRole)
	customRoles.Delete("/:id", ch.DeleteCustomRole)
}

func RegisterProfileRoutes(router fiber.Router, ph *ProfileHandler, m *middlewares.AuthMiddleware) {
	profile := router.Group("/profile", m.RequireAuth)

	profile.Get("/", ph.GetProfile)
	profile.Patch("/", ph.UpdateProfile)
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

	profile.Get("/preferences", ph.GetPreferences)
	profile.Patch("/preferences", ph.UpdatePreferences)

	profile.Get("/notifications", ph.GetNotifications)
	profile.Patch("/notifications", ph.UpdateNotifications)

	profile.Get("/sessions", ph.GetSessions)
	profile.Delete("/sessions/:id", ph.RevokeSession)

	profile.Get("/activity", ph.GetActivity)
}
