package dto

import (
	"encoding/json"
	"sage-backend/internal/shield/domain"
)

func SecurityEventToResponse(s *domain.SecurityEvent) *SecurityEventResponse {
	if s == nil {
		return nil
	}
	pid := ""
	if s.ParserID != nil {
		pid = s.ParserID.String()
	}
	sev := ""
	if s.Severity != nil {
		sev = string(*s.Severity)
	}
	return &SecurityEventResponse{
		ID:                s.ID.String(),
		SourceID:          s.SourceID.String(),
		Source:            s.Source,
		SourceEventID:     s.SourceEventID,
		EventType:         s.EventType,
		EventCategory:     string(s.EventCategory),
		Severity:          sev,
		ActorEmail:        s.ActorEmail,
		ActorUsername:     s.ActorUsername,
		IPAddress:         s.IPAddress,
		GeoCountry:        s.GeoCountry,
		GeoCity:           s.GeoCity,
		RawPayload:        s.RawPayload,
		NormalizedPayload: s.NormalizedPayload,
		ParserID:          &pid,
		ParseStatus:       string(s.ParseStatus),
		ParseErrors:       s.ParseErrors,
		OccurredAt:        s.OccurredAt,
		IngestedAt:        s.IngestedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

func DataSourceToResponse(d *domain.DataSource) *DataSourceResponse {
	if d == nil {
		return nil
	}
	meta := d.Metadata
	if meta == nil {
		meta = json.RawMessage{}
	}
	return &DataSourceResponse{
		ID:               d.ID.String(),
		Name:             d.Name,
		Description:      d.Description,
		Type:             d.Type,
		Provider:         d.Provider,
		Status:           string(d.Status),
		EventsToday:      d.EventsToday,
		TotalEvents:      d.TotalEvents,
		LastEventAt:      d.LastEventAt,
		LastSyncAt:       d.LastSyncAt,
		ErrorCount:       d.ErrorCount,
		DelayedByMinutes: d.DelayedByMinutes,
		Metadata:         meta,
		LastCheckpoint:   d.LastCheckpoint,
		LastCheckpointAt: d.LastCheckpointAt,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func ParserToResponse(p *domain.Parser) *ParserResponse {
	if p == nil {
		return nil
	}
	ownerStr := ""
	if p.OwnerUserID != nil {
		ownerStr = p.OwnerUserID.String()
	}
	return &ParserResponse{
		ID:              p.ID.String(),
		Name:            p.Name,
		Description:     p.Description,
		ParserType:      string(p.ParserType),
		Status:          string(p.Status),
		Tags:            p.Tags,
		Logic:           p.Logic,
		Mappings:        p.Mappings,
		EventsParsed24h: p.EventsParsed24h,
		ErrorRate:       p.ErrorRate,
		Owner:           &ownerStr,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func IngestionJobToResponse(i *domain.IngestionJob) *IngestionJobResponse {
	if i == nil {
		return nil
	}
	sid := ""
	if i.SourceID != nil {
		sid = i.SourceID.String()
	}
	return &IngestionJobResponse{
		ID:              i.ID.String(),
		OrganizationID:  i.OrganizationID.String(),
		SourceID:        &sid,
		Status:          string(i.Status),
		JobType:         string(i.JobType),
		EventsProcessed: i.EventsProcessed,
		EventsFailed:    i.EventsFailed,
		ErrorMessage:    i.ErrorMessage,
		Metadata:        i.Metadata,
		StartedAt:       i.StartedAt,
		CompletedAt:     i.CompletedAt,
		CreatedAt:       i.CreatedAt,
	}
}

func DataQualityScanToResponse(d *domain.DataQualityScan) *DataQualityScanResponse {
	if d == nil {
		return nil
	}
	return &DataQualityScanResponse{
		ID:                        d.ID.String(),
		Status:                    d.Status,
		QualityScore:              d.QualityScore,
		ParsingErrors:             d.ParsingErrors,
		MissingFieldsPercentage:   d.MissingFieldsPercentage,
		DuplicateEventsPercentage: d.DuplicateEventsPercentage,
		UnmappedLogsCount:         d.UnmappedLogsCount,
		AISummary:                 d.AISummary,
		StartedAt:                 d.StartedAt,
		CompletedAt:               d.CompletedAt,
		CreatedAt:                 d.CreatedAt,
	}
}

func DataQualityMetricToResponse(d *domain.DataQualitySourceMetric, sourceName string) *DataQualityBreakdownResponse {
	if d == nil {
		return nil
	}
	return &DataQualityBreakdownResponse{
		SourceID:                d.SourceID.String(),
		SourceName:              sourceName,
		ParsingErrors:           d.ParsingErrors,
		MissingFieldsPercentage: d.MissingFieldsPercentage,
		UnmappedEvents:          d.UnmappedEvents,
		DuplicatePercentage:     d.DuplicatePercentage,
		Status:                  string(d.Status),
	}
}

func DataQualitySuggestionToResponse(d *domain.DataQualitySuggestion) *SuggestionResponse {
	if d == nil {
		return nil
	}
	sid := ""
	if d.SourceID != nil {
		sid = d.SourceID.String()
	}
	pid := ""
	if d.ParserID != nil {
		pid = d.ParserID.String()
	}
	return &SuggestionResponse{
		ID:             d.ID.String(),
		SourceID:       &sid,
		ParserID:       &pid,
		Summary:        d.Summary,
		Recommendation: d.Recommendation,
		SuggestedFix:   d.SuggestedFix,
		Confidence:     d.Confidence,
		Status:         string(d.Status),
		CreatedAt:      d.CreatedAt,
		AppliedAt:      d.AppliedAt,
	}
}
