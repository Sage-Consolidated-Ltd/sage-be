package routes

import (
	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shield/handlers"
	shieldMiddlewares "sage-backend/internal/shield/middlewares"

	"github.com/gofiber/fiber/v2"
)

type LogsDataDeps struct {
	LDH *handlers.LogsDataHandler
	PH  *handlers.ParserHandler
	QH  *handlers.QualityHandler
	IH  *handlers.IntegrationHandler
	PLH *handlers.ParsedLogHandler
	UH  *handlers.UploadHandler
	M   *middlewares.AuthMiddleware
	RL  *shieldMiddlewares.RateLimiter
}

func Setup(
	app *fiber.App,
	ih *handlers.IntegrationHandler,
	qh *handlers.QualityHandler,
	ldh *handlers.LogsDataHandler,
	ph *handlers.ParserHandler,
	eh *handlers.EventHandler,
	uh *handlers.UploadHandler,
	plh *handlers.ParsedLogHandler,
	rl *shieldMiddlewares.RateLimiter,
	m *middlewares.AuthMiddleware) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to SAGE Shield SERVICE")
	})

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		return response.JSON(c, fiber.StatusOK, "API is Healthy", nil)
	})

	logsDataDeps := LogsDataDeps{
		LDH: ldh,
		PH:  ph,
		QH:  qh,
		IH:  ih,
		PLH: plh,
		UH:  uh,
		M:   m,
		RL:  rl,
	}

	RegisterLogsDataRoutes(v1, logsDataDeps)
	RegisterEventRoutes(v1, eh, m)

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "route not found", nil)
	})
}

func RegisterEventRoutes(router fiber.Router, eh *handlers.EventHandler, m *middlewares.AuthMiddleware) {
	events := router.Group("/events")
	events.Get("/logs", m.RequireAuth, eh.SearchLogs)
	events.Get("/logs/:id", m.RequireAuth, eh.GetLogDetail)
	events.Post("/logs/ingest", m.RequireAuth, eh.IngestLog)
	events.Post("/logs/bulk-ingest", m.RequireAuth, eh.BulkIngestLogs)
}

func RegisterLogsDataRoutes(router fiber.Router, deps LogsDataDeps) {
	logsData := router.Group("/integrations/logs-data")

	registerIngestionHealthRoutes(logsData, deps.LDH, deps.M)
	registerSourceRoutes(logsData, deps.LDH, deps.M)
	registerParserRoutes(logsData, deps.PH, deps.M)
	registerQualityRoutes(logsData, deps.QH, deps.M)
	registerUploadRoutes(logsData, deps.UH, deps.M, deps.RL)
	registerParsedLogRoutes(logsData, deps.PLH, deps.M)
}

func registerParsedLogRoutes(router fiber.Router, plh *handlers.ParsedLogHandler, m *middlewares.AuthMiddleware) {
	parsedLog := router.Group("/parsed-logs", m.RequireAuth)
	parsedLog.Get("/", plh.SearchParsedLogs)
}

func registerUploadRoutes(router fiber.Router, uh *handlers.UploadHandler, m *middlewares.AuthMiddleware, rl *shieldMiddlewares.RateLimiter) {
	upload := router.Group(
		"/upload",
		m.RequireAuth,
		rl.Handler(),
		shieldMiddlewares.RequestID(),
		shieldMiddlewares.Recover(logger.Default()),
		shieldMiddlewares.MaxBody(20*1024*1024), // 20MB
		shieldMiddlewares.Logger(logger.Default()),
	)

	upload.Post("/presign", uh.UploadLog)	 // upload log and get pre-signed URL
	upload.Post("/complete", uh.UploadComplete) // notify server that upload is complete
}

func registerIngestionHealthRoutes(router fiber.Router, ldh *handlers.LogsDataHandler, m *middlewares.AuthMiddleware) {
	ingestionHealth := router.Group("/ingestion-health", m.RequireAuth)
	ingestionHealth.Get("/", ldh.GetIngestionHealth)
	ingestionHealth.Post("/refresh", ldh.RefreshIngestionHealth)
	ingestionHealth.Get("/volume", ldh.GetIngestionVolume)
	ingestionHealth.Get("/notifications", ldh.GetIngestionNotifications)
	ingestionHealth.Get("/report", ldh.DownloadIngestionHealthReport)
}

func registerSourceRoutes(router fiber.Router, ldh *handlers.LogsDataHandler, m *middlewares.AuthMiddleware) {
	sources := router.Group("/sources", m.RequireAuth)
	sources.Get("/", ldh.ListSources)
	sources.Get("/:id", ldh.GetSource)
	sources.Post("/:id/sync", ldh.SyncSource)
	sources.Post("/:id/disconnect", ldh.DisconnectSource)
	sources.Get("/:id/logs", ldh.GetSourceLogs)
}

func registerParserRoutes(router fiber.Router, ph *handlers.ParserHandler, m *middlewares.AuthMiddleware) {
	parsers := router.Group("/parsers", m.RequireAuth)
	parsers.Get("/summary", ph.GetParserSummary)
	parsers.Get("/", ph.ListParsers)
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
	parsers.Get("/sample-logs", ph.ListSampleLogs)
}

func registerQualityRoutes(router fiber.Router, qh *handlers.QualityHandler, m *middlewares.AuthMiddleware) {
	quality := router.Group("/data-quality", m.RequireAuth)
	quality.Get("/", qh.GetDataQualitySummary)
	quality.Post("/scan", qh.RunDataQualityScan)
	quality.Get("/breakdown", qh.GetDataQualityBreakdown)
	quality.Get("/ai-analysis", qh.GetAIAnalysis)
	quality.Post("/apply-suggested-fix", qh.ApplySuggestedFix)
	quality.Post("/diff", qh.GetSuggestedFixDiff)
	quality.Get("/report", qh.DownloadDataQualityReport)
}