package middlewares

import (
	"errors"
	"fmt"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
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

	if sess.Get("pending_email_verification") != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Email verification required"})
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

func GetOrgID(c *fiber.Ctx) (uuid.UUID, error) {
	if config.Store == nil {
		return uuid.Nil, errors.New("session store not initialized")
	}

	sess, err := config.Store.Get(c)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get session: %w", err)
	}

	orgIDRaw := sess.Get("organizationID")
	if orgIDRaw == nil {
		return uuid.Nil, errors.New("organization ID missing in session")
	}

	orgIDStr, ok := orgIDRaw.(string)
	if !ok {
		return uuid.Nil, errors.New("organization ID is not a string")
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	return orgID, nil
}

func GetOrgIDStr(c *fiber.Ctx) string {
	sess, err := config.Store.Get(c)
	if err != nil {
		return ""
	}
	orgIDStr := sess.Get("organizationID")
	if orgIDStr == nil {
		return ""
	}
	return orgIDStr.(string)
}

func GetUserID(c *fiber.Ctx) uuid.UUID {
	sess, err := config.Store.Get(c)
	if err != nil {
		return uuid.Nil
	}
	userIDStr := sess.Get("userID")
	if userIDStr == nil {
		return uuid.Nil
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil
	}
	return userID
}

func GetUserIDStr(c *fiber.Ctx) string {
	sess, err := config.Store.Get(c)
	if err != nil {
		return ""
	}
	userIDStr := sess.Get("userID")
	if userIDStr == nil {
		return ""
	}
	return userIDStr.(string)
}

func GetRole(c *fiber.Ctx) string {
	sess, err := config.Store.Get(c)
	if err != nil {
		return ""
	}
	role := sess.Get("role")
	if role == nil {
		return ""
	}
	return role.(string)
}

func GetSessionInfo(c *fiber.Ctx) (string, string, string, bool) {
	orgID := GetOrgIDStr(c)
	userID := GetUserIDStr(c)
	role := GetRole(c)
	if orgID == "" {
		return "", "", "", false
	}
	return orgID, userID, role, true
}
