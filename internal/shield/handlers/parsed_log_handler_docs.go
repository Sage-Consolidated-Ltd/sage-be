package handlers

// @Summary Search parsed logs
// @Description Searches parsed logs within the authenticated user's organization. Supports two query styles: (1) a single Splunk-style "query" string combining field filters, raw JSON field filters, and free-text phrases (e.g. level=error "connection refused" raw.eventid=4625), or (2) individual flat parameters (level, q, source) for simpler integrations. If "query" is provided, it takes precedence and the flat filter parameters are ignored. Time range, cursor, and limit apply regardless of which style is used. Results are scoped strictly to the caller's organization and paginated via a timestamp-based cursor for efficient traversal of large result sets.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param query query string false "Splunk-style query string combining filters and free text, e.g. level=error \"connection refused\" raw.eventid=4625. Takes precedence over level/q/source if provided."
// @Param level query string false "Filter by log level (e.g. error, warning, info). Ignored if 'query' is provided."
// @Param q query string false "Free-text search against the log message body. Ignored if 'query' is provided."
// @Param source query string false "Filter by data source ID (UUID). Ignored if 'query' is provided."
// @Param from query string false "Only include logs at or after this timestamp (RFC3339, e.g. 2026-07-01T00:00:00Z)"
// @Param to query string false "Only include logs at or before this timestamp (RFC3339)"
// @Param cursor query string false "Pagination cursor from a previous response's next_cursor field (RFC3339 timestamp); returns logs older than this value"
// @Param limit query int false "Maximum number of logs to return per page (default: 50, max: 500)"
// @Success 200 {object} response.Response "List of matching parsed logs scoped to the caller's organization, with a cursor for the next page if more results exist"
// @Failure 400 {object} response.Response "Invalid or malformed query parameters"
// @Failure 401 {object} response.Response "Missing or invalid authentication"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /integrations/logs-data/parsed-logs [get]
func _SearchParsedLogs() {}
