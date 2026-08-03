package http

import (
	"sage-backend/internal/shared/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetUpRouter(
	router fiber.Router,
	ih *IntegrationHandler,
	qh *QualityHandler,
	ldh *LogsDataHandler,
	ph *ParserHandler,
	eh *EventHandler,
	m *middlewares.AuthMiddleware,
) {
	RegisterLogsDataRoutes(router, ldh, ph, qh, ih, m)
	RegisterEventRoutes(router, eh, m)
}

func RegisterLogsDataRoutes(
	router fiber.Router,
	ldh *LogsDataHandler,
	ph *ParserHandler,
	qh *QualityHandler,
	ih *IntegrationHandler,
	m *middlewares.AuthMiddleware,
) {
	integrations := router.Group("/integrations")
	if ih != nil {
		integrations.Post("/", m.RequireAuth, ih.IntegrateDataSource)
	}

	logsData := integrations.Group("/logs-data")
	if ldh != nil {
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
	}

	parsers := logsData.Group("/parsers")
	if ph != nil {
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
	}

	quality := logsData.Group("/data-quality", m.RequireAuth)
	if qh != nil {
		quality.Get("/", qh.GetDataQualitySummary)
		quality.Post("/scan", qh.RunDataQualityScan)
		quality.Get("/breakdown", qh.GetDataQualityBreakdown)
		quality.Get("/ai-analysis", qh.GetAIAnalysis)
		quality.Post("/apply-suggested-fix", qh.ApplySuggestedFix)
		quality.Post("/diff", qh.GetSuggestedFixDiff)
		quality.Get("/report", qh.DownloadDataQualityReport)
	}
}

func RegisterEventRoutes(router fiber.Router, eh *EventHandler, m *middlewares.AuthMiddleware) {
	if eh != nil {
		events := router.Group("/events")
		events.Get("/logs", m.RequireAuth, eh.SearchLogs)
		events.Get("/logs/:id", m.RequireAuth, eh.GetLogDetail)
		events.Post("/logs/ingest", m.RequireAuth, eh.IngestLog)
		events.Post("/logs/bulk-ingest", m.RequireAuth, eh.BulkIngestLogs)
	}
}
