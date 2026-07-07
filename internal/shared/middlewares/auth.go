package middlewares

import (
	"errors"
	"fmt"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthMiddleware struct {
	RedisClient *redis.Client
}


type RequestContext struct {
	UserID         string
	RequestID      string
	OrganizationID string
	RoleInOrganization string
}

func UserActivityKey(orgID, userID string) string {
	if orgID == "" {
		return fmt.Sprintf("activity:last_active:user:%s", userID)
	}

	return fmt.Sprintf("activity:last_active:org:%s:user:%s", orgID, userID)
}

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

	if am.RedisClient != nil {
		userID, _ := sess.Get("userID").(string)
		orgID, _ := sess.Get("activeOrgID").(string)
		if userID != "" {
			now := time.Now().UTC()
			sess.Set("last_active_at", now.Format(time.RFC3339Nano))
			_ = sess.Save()
			_ = am.RedisClient.Set(c.Context(), UserActivityKey(orgID, userID), now.Format(time.RFC3339Nano), 24*time.Hour).Err()
		}
	}
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

	orgIDRaw := sess.Get("activeOrgID")
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

func GetRequestContext(c *fiber.Ctx) RequestContext {
	requestID, _ := c.Locals("requestID").(string)

	return RequestContext{
		UserID:         GetUserIDStr(c),   // reads from session
		OrganizationID: GetOrgIDStr(c),    // reads from session
		RoleInOrganization: GetRoleInOrgStr(c), // reads from session
		RequestID:      requestID,
	}
}

func GetOrgIDStr(c *fiber.Ctx) string {
	sess, err := config.Store.Get(c)
	if err != nil {
		return ""
	}
	orgIDStr := sess.Get("activeOrgID")
	if orgIDStr == nil {
		return ""
	}
	return orgIDStr.(string)
}
func GetRoleInOrgStr(c *fiber.Ctx) string {
	sess, err := config.Store.Get(c)
	if err != nil {
		return ""
	}
	roleInOrg := sess.Get("roleInOrganization")
	if roleInOrg == nil {
		return ""
	}
	return roleInOrg.(string)
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

func GetSessionInfo(c *fiber.Ctx) (string, string, string, string, bool) {
	orgID := GetOrgIDStr(c)
	roleInOrg := GetRoleInOrgStr(c)
	userID := GetUserIDStr(c)
	role := GetRole(c)
	if orgID == "" {
		return "", "", "", "", false
	}
	return orgID, roleInOrg, userID, role, true
}
