package http

import (
	"github.com/gofiber/fiber/v2"
)

func SetUpAuthRouter(
	router fiber.Router,
	handler *AuthHandler,
	middleware *AdminAuthMiddleware,
) {
	auth := router.Group("/admin/auth")

	// Public auth endpoints
	auth.Post("/login", handler.Login)
	auth.Post("/verify-otp", handler.VerifyOTP)
	auth.Post("/resend-otp", handler.ResendOTP)

	// Authenticated admin endpoints
	auth.Get("/me", middleware.RequireAdminAuth(), handler.Me)
	auth.Post("/logout", middleware.RequireAdminAuth(), handler.Logout)
}
