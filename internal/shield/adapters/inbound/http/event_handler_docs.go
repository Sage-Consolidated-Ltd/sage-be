package http

import (
	_ "sage-backend/internal/shared/response"
)

// @Summary Search Logs (AST & Filtering)
// @Description Searches parsed logs across API-polled integrations (Okta, Entra) and S3 uploaded files. Supports AST query syntax using parameter 'q' (e.g., q='level=ERROR "unauthorized access" raw.ip=192.168.1.1').
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param q query string false "AST Search Query string (e.g., level=ERROR 'failed login' raw.user=johndoe)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 25, max: 200)"
// @Param source_id query string false "Filter by Data source ID"
// @Param event_type query string false "Filter by Event type"
// @Param severity query string false "Filter by Severity (low, medium, high, critical)"
// @Param actor_email query string false "Filter by Actor email"
// @Param ip_address query string false "Filter by IP address"
// @Param start_time query string false "Start timestamp (RFC3339 format)"
// @Param end_time query string false "End timestamp (RFC3339 format)"
// @Param search query string false "Free text fallback search"
// @Success 200 {object} response.Response
// @Router /events/logs [get]
func _SearchLogs() {}

// @Summary Get Log Detail
// @Description Returns full raw and normalized payload for one security event.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "Log Event ID"
// @Success 200 {object} response.Response
// @Router /events/logs/{id} [get]
func _GetLogDetail() {}

// @Summary Ingest Log Event
// @Description Ingests one raw security event, normalizes it, and stores it for investigation and detection.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body dto.IngestLogRequest true "Ingest Log Request"
// @Success 200 {object} response.Response
// @Router /logs/ingest [post]
func _IngestLog() {}

// @Summary Bulk Ingest Logs
// @Description Bulk ingests raw security events for a data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body dto.BulkIngestLogsRequest true "Bulk Ingest Logs Request"
// @Success 200 {object} response.Response
// @Router /logs/bulk-ingest [post]
func _BulkIngestLogs() {}
