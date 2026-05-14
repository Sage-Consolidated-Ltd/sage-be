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
	qh *handlers.QualityHandler,
	ldh *handlers.LogsDataHandler,
	ph *handlers.ParserHandler,
	eh *handlers.EventHandler,
	m *middlewares.AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE Shield SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	RegisterLogsDataRoutes(v1, ldh, ph, qh, ih, m)
	RegisterEventRoutes(v1, eh, m)

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}
func RegisterLogsDataRoutes(
	router fiber.Router,
	ldh *handlers.LogsDataHandler,
	ph *handlers.ParserHandler,
	qh *handlers.QualityHandler,
	ih *handlers.IntegrationHandler,
	m *middlewares.AuthMiddleware,
) {
	integrations := router.Group("/integrations")
	integrations.Post("/", m.RequireAuth, ih.IntegrateDataSource)

	logsData := integrations.Group("/logs-data")
	logsData.Get("/ingestion-health", m.RequireAuth, ldh.GetIngestionHealth)
	logsData.Post("/ingestion-health/refresh", m.RequireAuth, ldh.RefreshIngestionHealth)
	logsData.Get("/ingestion-health/volume", m.RequireAuth, ldh.GetIngestionVolume)
	logsData.Get("/ingestion-health/notifications", m.RequireAuth, ldh.GetIngestionNotifications)
	logsData.Get("/ingestion-health/report", m.RequireAuth, ldh.DownloadIngestionHealthReport)
	logsData.Get("/sources", m.RequireAuth, ldh.ListSources)
	logsData.Get("/sources/:id", m.RequireAuth, ldh.GetSource)
	logsData.Post("/sources/:id/sync", m.RequireAuth, ldh.SyncSource)
	logsData.Post("/sources/:id/disconnect", m.RequireAuth, ldh.DisconnectSource)
	logsData.Get("/sources/:id/logs", m.RequireAuth, ldh.GetSourceLogs)

	parsers := logsData.Group("/parsers")
	parsers.Get("/summary", m.RequireAuth, ph.GetParserSummary)
	parsers.Get("/", m.RequireAuth, ph.ListParsers)
	parsers.Get("/:id", m.RequireAuth, ph.GetParser)
	parsers.Post("/", m.RequireAuth, ph.CreateParser)
	parsers.Patch("/:id", m.RequireAuth, ph.UpdateParser)
	parsers.Post("/:id/test", m.RequireAuth, ph.TestParser)
	parsers.Post("/preview", m.RequireAuth, ph.PreviewParser)
	parsers.Post("/:id/enable", m.RequireAuth, ph.EnableParser)
	parsers.Post("/:id/disable", m.RequireAuth, ph.DisableParser)
	parsers.Post("/:id/validate", m.RequireAuth, ph.ValidateParser)
	parsers.Post("/validate", m.RequireAuth, ph.ValidateAllParsers)
	parsers.Post("/import", m.RequireAuth, ph.ImportParser)
	parsers.Get("/:id/export", m.RequireAuth, ph.ExportParser)
	parsers.Get("/sample-logs", m.RequireAuth, ph.ListSampleLogs)

	quality := logsData.Group("/data-quality", m.RequireAuth)
	quality.Get("/", qh.GetDataQualitySummary)
	quality.Post("/scan", qh.RunDataQualityScan)
	quality.Get("/breakdown", qh.GetDataQualityBreakdown)
	quality.Get("/ai-analysis", qh.GetAIAnalysis)
	quality.Post("/apply-suggested-fix", qh.ApplySuggestedFix)
	quality.Post("/diff", qh.GetSuggestedFixDiff)
	quality.Get("/report", qh.DownloadDataQualityReport)
}

func RegisterEventRoutes(router fiber.Router, eh *handlers.EventHandler, m *middlewares.AuthMiddleware) {
	events := router.Group("/events")
	events.Get("/logs", m.RequireAuth, eh.SearchLogs)
	events.Get("/logs/:id", m.RequireAuth, eh.GetLogDetail)
	events.Post("/logs/ingest", m.RequireAuth, eh.IngestLog)
	events.Post("/logs/bulk-ingest", m.RequireAuth, eh.BulkIngestLogs)
}
