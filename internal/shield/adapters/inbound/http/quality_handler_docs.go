package http

import (
	_ "sage-backend/internal/shared/response"
	_ "sage-backend/internal/shield/ports/dto"
)

// @Summary Get Data Quality Summary
// @Description Returns overall data quality metrics including score, parsing errors, missing fields, unmapped logs, and duplicate events.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality [get]
func _GetDataQualitySummary() {}

// @Summary Run Data Quality Scan
// @Description Queues a data quality scan to detect parsing errors, missing fields, duplicates, and unmapped logs.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality/scan [post]
func _RunDataQualityScan() {}

// @Summary Get Data Quality Breakdown
// @Description Returns data quality metrics per data source.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Quality status"
// @Param source_id query string false "Data source ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality/breakdown [get]
func _GetDataQualityBreakdown() {}

// @Summary Get Data Quality AI Analysis
// @Description Returns AI-generated analysis and suggested fixes for data quality issues.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality/ai-analysis [get]
func _GetAIAnalysis() {}

// @Summary Apply Suggested Data Quality Fix
// @Description Applies an AI-suggested parser or field mapping fix to improve data quality.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body dto.ApplySuggestedFixRequest true "Apply Suggested Fix Request"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality/apply-suggested-fix [post]
func _ApplySuggestedFix() {}

// @Summary Preview Suggested Fix Diff
// @Description Returns before and after diff for an AI-suggested parser or mapping fix.
// @Tags Logs & Data
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param suggestion_id query string true "Suggestion ID"
// @Param parser_id query string true "Parser ID"
// @Success 200 {object} response.Response
// @Router /integrations/logs-data/data-quality/diff [get]
func _GetSuggestedFixDiff() {}

// @Summary Download Data Quality Report
// @Description Downloads data quality report in csv, pdf, or json format.
// @Tags Logs & Data
// @Accept json
// @Produce application/octet-stream
// @Security SessionAuth
// @Param format query string false "Report format"
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Success 200 {file} file
// @Router /integrations/logs-data/data-quality/report [get]
func _DownloadDataQualityReport() {}
