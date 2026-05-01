package routes

import (
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App,
	ih *handlers.IntegrationHandler,
	ldh *handlers.LogsDataHandler,
	ph *handlers.ParserHandler,
	qh *handlers.QualityHandler,
	m *middlewares.AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE Shield SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	RegisterIntegrationRoutes(v1, ih, m)
	RegisterLogsDataRoutes(v1, ldh, m)
	RegisterParserRoutes(v1, ph, m)
	RegisterQualityRoutes(v1, qh, m)

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}

func RegisterIntegrationRoutes(router fiber.Router, ih *handlers.IntegrationHandler, m *middlewares.AuthMiddleware) {
	// integration := router.Group("/integrations")
	// integration.Post("/", m.RequireAuth, ih.CreateIntegration) // Placeholder as per repo state
}

func RegisterLogsDataRoutes(router fiber.Router, ldh *handlers.LogsDataHandler, m *middlewares.AuthMiddleware) {
	logsData := router.Group("/logs-data")
	
	// Ingestion Health
	logsData.Get("/ingestion-health", ldh.GetIngestionHealth)
	logsData.Post("/ingestion-health/refresh", ldh.RefreshIngestionHealth)
	logsData.Get("/ingestion-health/volume", ldh.GetIngestionVolume)
	logsData.Get("/ingestion-health/notifications", ldh.GetIngestionNotifications)
	logsData.Get("/ingestion-health/report", ldh.DownloadIngestionHealthReport)

	// Sources
	logsData.Get("/sources", ldh.ListSources)
	logsData.Get("/sources/:id", ldh.GetSource)
	logsData.Post("/sources/:id/sync", ldh.SyncSource)
	logsData.Post("/sources/:id/disconnect", ldh.DisconnectSource)
	logsData.Get("/sources/:id/logs", ldh.GetSourceLogs)
}

func RegisterParserRoutes(router fiber.Router, ph *handlers.ParserHandler, m *middlewares.AuthMiddleware) {
	parsers := router.Group("/logs-data/parsers")
	
	parsers.Get("/summary", ph.GetParserSummary)
	parsers.Get("/", ph.ListParsers)
	parsers.Get("/sample-logs", ph.ListSampleLogs)
	parsers.Get("/:id", ph.GetParser)
	parsers.Post("/", ph.CreateParser)
	parsers.Patch("/:id", ph.UpdateParser)
	parsers.Post("/:id/test", ph.TestParser)
	parsers.Post("/preview", ph.PreviewParser)
	parsers.Post("/:id/enable", ph.EnableParser)
	parsers.Post("/:id/disable", ph.DisableParser)
	parsers.Post("/:id/validate", ph.ValidateParser)
	parsers.Post("/validate", ph.ValidateAllParsers)
	parsers.Post("/import", ph.ImportParser)
	parsers.Get("/:id/export", ph.ExportParser)
}

func RegisterQualityRoutes(router fiber.Router, qh *handlers.QualityHandler, m *middlewares.AuthMiddleware) {
	quality := router.Group("/logs-data/data-quality")
	
	quality.Get("/", qh.GetDataQualitySummary)
	quality.Post("/scan", qh.RunDataQualityScan)
	quality.Get("/breakdown", qh.GetDataQualityBreakdown)
	quality.Get("/ai-analysis", qh.GetAIAnalysis)
	quality.Post("/apply-suggested-fix", qh.ApplySuggestedFix)
	quality.Get("/diff", qh.GetSuggestedFixDiff)
	quality.Get("/report", qh.DownloadDataQualityReport)
}
