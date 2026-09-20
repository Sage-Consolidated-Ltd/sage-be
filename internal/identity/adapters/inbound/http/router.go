package http

import (
	"sage-backend/internal/shared/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetUpRouter(router fiber.Router, ah *AuthHandler, ph *ProfileHandler, m *middlewares.AuthMiddleware) {
	RegisterAuthRoutes(router, ah, m)
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
	auth.Post("/logout", ah.Logout)
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

	profile.Post("/change-password", ph.ChangePassword)
	profile.Post("/backup-email", ph.ConfigureBackupEmail)
	profile.Delete("/me", ph.DeleteAccount)
	profile.Delete("/account", ph.DeleteAccount)
}
