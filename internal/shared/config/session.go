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
	storageConfig := redis.Config{}
	if cfg.RedisDbUrl != "" {
		storageConfig.URL = cfg.RedisDbUrl
	} else {
		storageConfig.Host = "localhost"
		storageConfig.Port = 6379
	}
	storage := redis.New(storageConfig)
	var sameSite string
	if cfg.APP_ENV == "production" || cfg.APP_ENV == "test" || cfg.APP_ENV == "testing" {
		sameSite = "Lax"
	} else {
		sameSite = "None"
	}

	cookieDomain := cfg.CookieDomain
	if cookieDomain == "localhost" || cookieDomain == "127.0.0.1" {
		cookieDomain = ""
	}

	Store = session.New(session.Config{
		Storage:        storage,
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:session_id",
		CookieHTTPOnly: true,
		CookieSecure:   cfg.APP_ENV == "production" || cfg.APP_ENV == "staging",
		CookieSameSite: sameSite,
		CookieDomain:   cookieDomain,
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
	sess.Set("authenticated", true)
	sess.Delete("pending_email_verification")
	sess.Delete("pending_2fa")

	if err := sess.Save(); err != nil {
		return nil, fmt.Errorf("could not save session: %w", err)
	}

	return sess, nil
}
