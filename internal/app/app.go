package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sage-backend/internal/shared/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// App manages the lifecycle of an HTTP application instance.
type App struct {
	Port  string
	Fiber *fiber.App
}

// NewFiberApp initializes a pre-configured Fiber instance with global middlewares.
func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder: func(v interface{}) ([]byte, error) {
			buf := &bytes.Buffer{}
			encoder := json.NewEncoder(buf)
			encoder.SetEscapeHTML(false)
			err := encoder.Encode(v)
			return bytes.TrimRight(buf.Bytes(), "\n"), err
		},
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0/0"},
		BodyLimit:               10 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, session_id, X-Session-ID",
	}))
	app.Use(recover.New())

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	return app
}

// Run starts the Fiber HTTP server and handles graceful shutdown signals.
func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		if err := a.Fiber.Listen(fmt.Sprintf(":%s", a.Port)); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	case <-quit:
		logger.Info("Shutting down gracefully...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Fiber.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("Shutdown complete")
	return nil
}
