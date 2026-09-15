package http

import (
	"errors"
	"strings"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/inbound"
	"sage-backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
)

const (
	AdminSessionCookie = "admin_session"
	LocalAdminSession  = "admin_session"
	LocalAdminID       = "admin_id"
	LocalAdminRole     = "admin_role"
)

type AdminAuthMiddleware struct {
	authUseCase inbound.AuthUseCase
}

func NewAdminAuthMiddleware(authUseCase inbound.AuthUseCase) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{authUseCase: authUseCase}
}

func (m *AdminAuthMiddleware) RequireAdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies(AdminSessionCookie)
		if sessionID == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				sessionID = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if sessionID == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Administrator authentication required", nil)
		}

		clientIP := c.IP()
		session, err := m.authUseCase.ValidateSession(c.Context(), sessionID, clientIP)
		if err != nil {
			if errors.Is(err, domain.ErrIPMismatch) {
				m.clearCookie(c)
				return response.Error(c, fiber.StatusUnauthorized, "Session invalidated due to IP change. Please log in again.", nil)
			}
			m.clearCookie(c)
			return response.Error(c, fiber.StatusUnauthorized, "Administrator session expired or invalid", nil)
		}

		c.Locals(LocalAdminSession, session)
		c.Locals(LocalAdminID, session.AdminID)
		c.Locals(LocalAdminRole, session.Role)

		return c.Next()
	}
}

func (m *AdminAuthMiddleware) RequireSuperAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(LocalAdminRole).(string)
		if !ok || role != string(domain.RoleSuperAdmin) {
			return response.Error(c, fiber.StatusForbidden, "Super Admin privileges required", nil)
		}
		return c.Next()
	}
}

func (m *AdminAuthMiddleware) clearCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     AdminSessionCookie,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})
}
