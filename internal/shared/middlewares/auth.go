package middlewares

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type AuthMiddleware struct{}

func (am *AuthMiddleware) RequireAuth(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	isVerifying := c.Path() == "/api/v1/auth/verify-2fa"

	if sess.Get("pending_2fa") != nil && !isVerifying {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "2FA verification required"})
	}

	c.Locals("session", sess)
	return c.Next()
}

func (am *AuthMiddleware) RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if sess, ok := c.Locals("session").(*session.Session); !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		} else if ok {
			userRole, _ := sess.Get("role").(string)
			if role == userRole {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, "Forbidden", nil)
	}
}
