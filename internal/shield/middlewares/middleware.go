package middlewares

import (
	"sage-backend/internal/shared/logger"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Chain(mw ...fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, h := range mw {
			if err := h(c); err != nil {
				return err
			}
			// If a handler short-circuited (e.g. rate limiter), stop the chain.
			if c.Response().StatusCode() >= 400 {
				return nil
			}
		}
		return c.Next()
	}
}
func Logger(log *logger.Logger) fiber.Handler {
	if log == nil {
		log = logger.Default()
	}
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			zap.String("ip", realIP(c)),
			zap.String("request_id", getRequestID(c)),
			zap.Int("bytes_in", len(c.Body())),
			zap.Int("bytes_out", len(c.Response().Body())),
		}

		switch {
		case status >= 500:
			log.Error("request", fields...)
		case status >= 400:
			log.Warn("request", fields...)
		default:
			log.Info("request", fields...)
		}

		return err
	}
}
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(fiber.HeaderXRequestID)
		if id == "" {
			id = generateID()
		}
		c.Set(fiber.HeaderXRequestID, id)
		c.Locals("requestID", id)
		return c.Next()
	}
}

func Recover(log *logger.Logger) fiber.Handler {
	if log == nil {
		log = logger.Default()
	}
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.String("ip", realIP(c)),
					zap.String("request_id", getRequestID(c)),
				)
				err = c.Status(fiber.StatusInternalServerError).
					JSON(fiber.Map{
						"code":    "INTERNAL_ERROR",
						"message": "an unexpected error occurred",
					})
			}
		}()
		return c.Next()
	}
}

func MaxBody(limit int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(c.Body()) > limit {
			return c.Status(fiber.StatusRequestEntityTooLarge).
				JSON(fiber.Map{
					"code":    "BODY_TOO_LARGE",
					"message": "request body exceeds limit",
				})
		}
		return c.Next()
	}
}
func realIP(c *fiber.Ctx) string {
	if fwd := c.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
	}
	if real := c.Get("X-Real-IP"); real != "" {
		return real
	}
	return c.IP()
}

func getRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestID").(string); ok {
		return id
	}
	return ""
}