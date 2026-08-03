package http

import (
	_ "sage-backend/internal/shared/response"
)

// @Summary Get Ingestion Health Summary
// @Description Returns ingestion health metrics including total events ingested, active sources, delayed sources, and ingestion errors.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/ingestion-health [get]
func _GetIngestionHealth() {}

// @Summary Refresh Ingestion Health
// @Description Queues a background job to refresh ingestion health metrics.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/ingestion-health/refresh [post]
func _RefreshIngestionHealth() {}

// @Summary List Data Sources
// @Description Returns connected data sources with ingestion status, events today, last event time, and health status.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by source name or description"
// @Param type query string false "Source type"
// @Param status query string false "Source status"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/sources [get]
func _ListSources() {}

// @Summary Get Data Source Detail
// @Description Returns detailed health and metadata for a single data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/sources/{id} [get]
func _GetSource() {}

// @Summary Sync Data Source
// @Description Manually queues a sync job for a data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/sources/{id}/sync [post]
func _SyncSource() {}

// @Summary Disconnect Data Source
// @Description Disconnects a data source and stops future ingestion while preserving historical logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Data Source ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/sources/{id}/disconnect [post]
func _DisconnectSource() {}

// @Summary View Source Logs
// @Description Returns logs ingested from a specific data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Data Source ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param event_type query string false "Event type"
// @Param severity query string false "Severity"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/sources/{id}/logs [get]
func _GetSourceLogs() {}

// @Summary Get Ingestion Volume Over Time
// @Description Returns time-series ingestion volume for charts, optionally filtered by source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Param interval query string false "Aggregation interval"
// @Param source_id query string false "Data source ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/ingestion-health/volume [get]
func _GetIngestionVolume() {}

// @Summary Get Ingestion Notifications
// @Description Returns ingestion warnings and AI insights related to source delays, drops, parsing errors, and schema changes.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/ingestion-health/notifications [get]
func _GetIngestionNotifications() {}

// @Summary Download Ingestion Health Report
// @Description Downloads ingestion health report in csv, pdf, or json format.
// @Tags Logs & Data
// @Accept json
// @Produce application/octet-stream
// @Security SessionAuth
// @Param format query string false "Report format"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {file} file
// @Router /integrations/logs-data/ingestion-health/report [get]
func _DownloadIngestionHealthReport() {}
