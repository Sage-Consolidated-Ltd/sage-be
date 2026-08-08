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
	LogFileID   uuid.UUID
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
	TotalEvents     int
	ThreatsDetected int
	ConfirmedBoth   int
	MLOnly          int
	RuleOnly        int
	ThreatRatePct   float64
}

type AnalysisOutcome struct {
	Status string
	Detail string
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
