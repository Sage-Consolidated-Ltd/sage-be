package models

import (
	"bytes"
	"sage-backend/internal/shared/db"
	"time"

	"github.com/google/uuid"
)

type AnalysisRequestType string

const (
	AnalysisRequestTypeFile AnalysisRequestType = "file"
	AnalysisRequestTypeJson AnalysisRequestType = "json"
)

type CreateAnalysisParams struct {
	LogFileID      *uuid.UUID
	JsonInputID    *uuid.UUID
	RequestType    AnalysisRequestType // "file" | "json"
	LogType        FileClass
	Approach       string
	Overall        string
	Summary        db.GenericJSON[AnalysisSummary]
	Outcome        db.GenericJSON[AnalysisOutcome]
	Threats        []Threat
	OrganizationID uuid.UUID
}

type SubmitLogFileInput struct {
	LogFileID      uuid.UUID
	S3Key          string
	FileClass      FileClass
	SourceType     string
	SourceID	   *uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

type SubmitLogFileForAnalysis struct {
	LogFileID uuid.UUID
	FileClass FileClass 
	OrganizationID uuid.UUID
	UserID uuid.UUID
	FileReader *bytes.Reader
	FileName string
	S3Key string
}

type SubmitLogFileResult struct {
	JobID       string    `json:"job_id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type AnalysisResult struct {
	ID uuid.UUID `db:"id" json:"id"`
	LogFileID uuid.UUID       `db:"log_file_id" json:"log_file_id"`
	LogType   string          `db:"log_type"    json:"log_type"`
	JsonInputID *uuid.UUID      `db:"json_input_id" json:"json_input_id"`
	RequestType AnalysisRequestType `db:"request_type" json:"request_type"`
	Approach  string          `db:"approach"    json:"approach"`
	Overall   string          `db:"overall"     json:"overall"`
	Summary   db.GenericJSON[AnalysisSummary] `db:"summary"     json:"summary"`
	Outcome   db.GenericJSON[AnalysisOutcome] `db:"outcome"     json:"outcome"`
	Threats   []Threat        `db:"-"           json:"threats"`
	CreatedAt time.Time       `db:"created_at"  json:"created_at"`
}

type AnalysisSummary struct {
	TotalEvents      int     `json:"total_events"`
	ThreatsDetected  int     `json:"threats_detected"`
	ConfirmedBoth    int     `json:"confirmed_both"`
	MLOnly           int     `json:"ml_only"`
	RuleOnly         int     `json:"rule_only"`
	ThreatRatePct    float64 `json:"threat_rate_pct"`
}

type AnalysisOutcome struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type Threat struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	AnalysisID     uuid.UUID `db:"analysis_id"     json:"analysis_id"`
	Source         string    `db:"source"          json:"source"`
	Title          string    `db:"title"           json:"title"`
	Category       string    `db:"category"        json:"category"`
	Severity       string    `db:"severity"        json:"severity"`
	Mitre          string    `db:"mitre"           json:"mitre"`
	EventCount     int       `db:"event_count"     json:"event_count"`
	TimeRange      string    `db:"time_range"      json:"time_range"`
	WhatHappened   string    `db:"what_happened"   json:"what_happened"`
	Evidence       []string  `db:"evidence"        json:"evidence"`
	Recommendation string    `db:"recommendation"  json:"recommendation"`
}

type CheckHealthResponse struct {
	Status string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}