package config

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
)

type SessionParam struct {
	ID             string
	Role           string
	Email          string
	OrganizationId string
}

var Store *session.Store

func InitSessionStore(cfg *BaseConfig) {
	storage := redis.New(redis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	})
	Store = session.New(session.Config{
		Storage:        storage,
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:session_id",
		CookieHTTPOnly: true,
		CookieSecure:   cfg.APP_ENV == "production",
		CookieSameSite: "Lax",
		// CookieDomain: "localhost",
	})
}

func SetSession(c *fiber.Ctx, param SessionParam) (*session.Session, error) {
	sess, err := Store.Get(c)
	if err != nil {
		return nil, fmt.Errorf("could not get session: %w", err)
	}

	sess.Set("userID", param.ID)
	sess.Set("role", param.Role)
	sess.Set("email", param.Email)
	sess.Set("organizationID", param.OrganizationId)

	if err := sess.Save(); err != nil {
		return nil, fmt.Errorf("could not save session: %w", err)
	}

	return sess, nil
}
