package domain

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
	RequestType    AnalysisRequestType
	LogType        FileClass
	Approach       string
	Overall        string
	Summary        db.GenericJSON[AnalysisSummary]
	Outcome        db.GenericJSON[AnalysisOutcome]
	Threats        []Threat
	OrganizationID uuid.UUID
}

type SubmitLogFileForAnalysis struct {
	LogFileID      uuid.UUID
	FileClass      FileClass
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	FileReader     *bytes.Reader
	FileName       string
	S3Key          string
}

type AnalysisResult struct {
	ID          uuid.UUID
	LogFileID   *uuid.UUID
	LogType     string
	JsonInputID *uuid.UUID
	RequestType AnalysisRequestType
	Approach    string
	Overall     string
	Summary     db.GenericJSON[AnalysisSummary]
	Outcome     db.GenericJSON[AnalysisOutcome]
	Threats     []Threat
	CreatedAt   time.Time
}

type AnalysisSummary struct {
	TotalEvents     int     `json:"total_events"`
	ThreatsDetected int     `json:"threats_detected"`
	ConfirmedBoth   int     `json:"confirmed_both"`
	MLOnly          int     `json:"ml_only"`
	RuleOnly        int     `json:"rule_only"`
	ThreatRatePct   float64 `json:"threat_rate_pct"`
}

type AnalysisOutcome struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type Threat struct {
	ID             uuid.UUID
	AnalysisID     uuid.UUID
	Source         string
	Title          string
	Category       string
	Severity       string
	Mitre          string
	EventCount     int
	TimeRange      string
	WhatHappened   string
	Evidence       []string
	Recommendation string
}
