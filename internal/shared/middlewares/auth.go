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

	userID := sess.Get("userID")
	if userID == nil || userID == "" {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	authenticated, _ := sess.Get("authenticated").(bool)
	if !authenticated {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	isVerifying2FA := c.Path() == "/api/v1/auth/verify-2fa"
	if sess.Get("pending_2fa") != nil && !isVerifying2FA {
		return response.Error(c, fiber.StatusForbidden, "2FA verification required", nil)
	}

	isVerifyingEmail := c.Path() == "/api/v1/auth/verify-email" || c.Path() == "/api/v1/auth/resend-verification"
	if sess.Get("pending_email_verification") != nil && !isVerifyingEmail {
		return response.Error(c, fiber.StatusForbidden, "Email verification required", nil)
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
	if config.Store == nil {
		if orgIDVal := c.Locals("orgID"); orgIDVal != nil {
			if id, ok := orgIDVal.(uuid.UUID); ok {
				return id.String()
			}
			if s, ok := orgIDVal.(string); ok {
				return s
			}
		}
		return ""
	}
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
	s := GetUserIDStr(c)
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func GetUserIDStr(c *fiber.Ctx) string {
	if config.Store == nil {
		if uidVal := c.Locals("userID"); uidVal != nil {
			if id, ok := uidVal.(uuid.UUID); ok {
				return id.String()
			}
			if s, ok := uidVal.(string); ok {
				return s
			}
		}
		return ""
	}
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
	if config.Store == nil {
		if roleVal := c.Locals("role"); roleVal != nil {
			if s, ok := roleVal.(string); ok {
				return s
			}
		}
		return ""
	}
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

type RequestContext struct {
	UserID             string
	RequestID          string
	OrganizationID     string
	RoleInOrganization string
}

func GetRequestContext(c *fiber.Ctx) RequestContext {
	requestID, _ := c.Locals("requestID").(string)

	return RequestContext{
		UserID:             GetUserIDStr(c),
		OrganizationID:     GetOrgIDStr(c),
		RoleInOrganization: GetRole(c),
		RequestID:          requestID,
	}
}

