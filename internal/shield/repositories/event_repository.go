package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
)

type SecurityEventRepositoryInt interface {
	CreateEvent(ctx context.Context, event *models.SecurityEvent) error
	BulkCreateEvents(ctx context.Context, events []*models.SecurityEvent) error
	GetEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.SecurityEvent, error)
	SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.SecurityEvent, int, error)
	GetEventsBySource(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.SecurityEvent, int, error)
	UpdateParseStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status types.ParseStatus, errors []map[string]interface{}, normalized map[string]interface{}) error
	GetEventsByParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, limit int) ([]*models.SecurityEvent, error)
	GetEventVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error)
	GetEventCountInWindow(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time) (int64, error)
}

type SecurityEventRepository struct {
	db *db.DB
}

func NewSecurityEventRepository(db *db.DB) SecurityEventRepositoryInt {
	return &SecurityEventRepository{db: db}
}

const (
	INSERT_EVENT = `
		INSERT INTO security_events (
			organization_id, source_id, parser_id, source_event_id, source,
			event_type, event_category, severity, actor_email, actor_username,
			ip_address, geo_country, geo_city, raw_payload, normalized_payload,
			parse_status, parse_errors, occurred_at, ingested_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW()
		) RETURNING id, created_at
	`
	BULK_INSERT_EVENTS = `
		INSERT INTO security_events (
			organization_id, source_id, parser_id, source_event_id, source,
			event_type, event_category, severity, actor_email, actor_username,
			ip_address, geo_country, geo_city, raw_payload, normalized_payload,
			parse_status, parse_errors, occurred_at, ingested_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
	`
	GET_EVENT     = `SELECT * FROM security_events WHERE id = $1 AND organization_id = $2`
	SEARCH_EVENTS = `
		SELECT * FROM security_events
		WHERE organization_id = $1
			AND ($2::uuid IS NULL OR source_id = $2)
			AND ($3::varchar IS NULL OR source = $3)
			AND ($4::varchar IS NULL OR event_type = $4)
			AND ($5::varchar IS NULL OR event_category = $5)
			AND ($6::varchar IS NULL OR severity = $6)
			AND ($7::varchar IS NULL OR actor_email ILIKE '%' || $7 || '%')
			AND ($8::varchar IS NULL OR ip_address ILIKE '%' || $8 || '%')
			AND ($9::timestamptz IS NULL OR occurred_at >= $9)
			AND ($10::timestamptz IS NULL OR occurred_at <= $10)
			AND ($11::varchar IS NULL OR (event_type ILIKE '%' || $11 || '%' OR source ILIKE '%' || $11 || '%'))
		ORDER BY occurred_at DESC
		LIMIT $12 OFFSET $13
	`
	COUNT_EVENTS = `
		SELECT COUNT(*) FROM security_events
		WHERE organization_id = $1
			AND ($2::uuid IS NULL OR source_id = $2)
			AND ($3::varchar IS NULL OR source = $3)
			AND ($4::varchar IS NULL OR event_type = $4)
			AND ($5::varchar IS NULL OR event_category = $5)
			AND ($6::varchar IS NULL OR severity = $6)
			AND ($7::varchar IS NULL OR actor_email ILIKE '%' || $7 || '%')
			AND ($8::varchar IS NULL OR ip_address ILIKE '%' || $8 || '%')
			AND ($9::timestamptz IS NULL OR occurred_at >= $9)
			AND ($10::timestamptz IS NULL OR occurred_at <= $10)
			AND ($11::varchar IS NULL OR (event_type ILIKE '%' || $11 || '%' OR source ILIKE '%' || $11 || '%'))
	`
	GET_EVENTS_BY_SOURCE = `
		SELECT * FROM security_events
		WHERE organization_id = $1 AND source_id = $2
			AND ($3::varchar IS NULL OR event_type = $3)
			AND ($4::varchar IS NULL OR severity = $4)
			AND ($5::timestamptz IS NULL OR occurred_at >= $5)
			AND ($6::timestamptz IS NULL OR occurred_at <= $6)
		ORDER BY occurred_at DESC
		LIMIT $7 OFFSET $8
	`
	COUNT_EVENTS_BY_SOURCE = `
		SELECT COUNT(*) FROM security_events
		WHERE organization_id = $1 AND source_id = $2
			AND ($3::varchar IS NULL OR event_type = $3)
			AND ($4::varchar IS NULL OR severity = $4)
			AND ($5::timestamptz IS NULL OR occurred_at >= $5)
			AND ($6::timestamptz IS NULL OR occurred_at <= $6)
	`
	UPDATE_PARSE_STATUS = `
		UPDATE security_events SET
			parse_status = $2,
			parse_errors = $3,
			normalized_payload = COALESCE($4, normalized_payload),
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $5
	`
	GET_EVENTS_BY_PARSER = `SELECT * FROM security_events WHERE parser_id = $1 AND organization_id = $2 AND parse_status = $3 LIMIT $4`
)

func (r *SecurityEventRepository) CreateEvent(ctx context.Context, event *models.SecurityEvent) error {
	var id uuid.UUID
	var createdAt time.Time
	raw := event.RawPayload
	if raw == nil {
		raw = make(map[string]interface{})
	}
	norm := event.NormalizedPayload
	if norm == nil {
		norm = make(map[string]interface{})
	}
	parseErrs := event.ParseErrors
	if parseErrs == nil {
		parseErrs = make([]map[string]interface{}, 0)
	}
	err := r.db.QueryRowContext(
		ctx, INSERT_EVENT,
		event.OrganizationID, event.SourceID, event.ParserID, event.SourceEventID, event.Source,
		event.EventType, event.EventCategory, event.Severity, event.ActorEmail, event.ActorUsername,
		event.IPAddress, event.GeoCountry, event.GeoCity, raw, norm,
		event.ParseStatus, parseErrs, event.OccurredAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	event.ID = id
	event.CreatedAt = createdAt
	return nil
}

func (r *SecurityEventRepository) BulkCreateEvents(ctx context.Context, events []*models.SecurityEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, BULK_INSERT_EVENTS)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		raw := event.RawPayload
		if raw == nil {
			raw = make(map[string]interface{})
		}
		norm := event.NormalizedPayload
		if norm == nil {
			norm = make(map[string]interface{})
		}
		parseErrs := event.ParseErrors
		if parseErrs == nil {
			parseErrs = make([]map[string]interface{}, 0)
		}
		_, err := stmt.ExecContext(
			ctx,
			event.OrganizationID, event.SourceID, event.ParserID, event.SourceEventID, event.Source,
			event.EventType, event.EventCategory, event.Severity, event.ActorEmail, event.ActorUsername,
			event.IPAddress, event.GeoCountry, event.GeoCity, raw, norm,
			event.ParseStatus, parseErrs, event.OccurredAt,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *SecurityEventRepository) GetEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.SecurityEvent, error) {
	var event models.SecurityEvent
	err := r.db.GetContext(ctx, &event, GET_EVENT, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("EVENT NOT FOUND")
		}
		return nil, err
	}
	return &event, nil
}

func (r *SecurityEventRepository) SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.SecurityEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	sourceID := uuid.Nil
	if sid, ok := filters["source_id"].(uuid.UUID); ok && sid != uuid.Nil {
		sourceID = sid
	} else if sidStr, ok := filters["source_id"].(string); ok && sidStr != "" {
		if uid, err := uuid.Parse(sidStr); err == nil {
			sourceID = uid
		}
	}
	source := ""
	if s, ok := filters["source"].(string); ok {
		source = s
	}
	eventType := ""
	if et, ok := filters["event_type"].(string); ok {
		eventType = et
	}
	category := ""
	if c, ok := filters["event_category"].(string); ok {
		category = c
	}
	severity := ""
	if s, ok := filters["severity"].(string); ok {
		severity = s
	}
	actorEmail := ""
	if a, ok := filters["actor_email"].(string); ok {
		actorEmail = a
	}
	ipAddress := ""
	if ip, ok := filters["ip_address"].(string); ok {
		ipAddress = ip
	}
	startTime := time.Time{}
	if st, ok := filters["start_time"].(time.Time); ok && !st.IsZero() {
		startTime = st
	}
	endTime := time.Time{}
	if et, ok := filters["end_time"].(time.Time); ok && !et.IsZero() {
		endTime = et
	}
	search := ""
	if s, ok := filters["search"].(string); ok {
		search = s
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_EVENTS,
		orgID, sourceID, source, eventType, category, severity, actorEmail, ipAddress, startTime, endTime, search)
	if err != nil {
		return nil, 0, err
	}

	var events []*models.SecurityEvent
	err = r.db.SelectContext(ctx, &events, SEARCH_EVENTS,
		orgID, sourceID, source, eventType, category, severity, actorEmail, ipAddress,
		startTime, endTime, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *SecurityEventRepository) GetEventsBySource(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.SecurityEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	eventType := ""
	if et, ok := filters["event_type"].(string); ok {
		eventType = et
	}
	severity := ""
	if s, ok := filters["severity"].(string); ok {
		severity = s
	}
	startTime := time.Time{}
	if st, ok := filters["start_time"].(time.Time); ok && !st.IsZero() {
		startTime = st
	}
	endTime := time.Time{}
	if et, ok := filters["end_time"].(time.Time); ok && !et.IsZero() {
		endTime = et
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_EVENTS_BY_SOURCE,
		orgID, sourceID, eventType, severity, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}

	var events []*models.SecurityEvent
	err = r.db.SelectContext(ctx, &events, GET_EVENTS_BY_SOURCE,
		orgID, sourceID, eventType, severity, startTime, endTime, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *SecurityEventRepository) UpdateParseStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status types.ParseStatus, errors []map[string]interface{}, normalized map[string]interface{}) error {
	_, err := r.db.ExecContext(
		ctx, UPDATE_PARSE_STATUS,
		id, status, errors, normalized, orgID,
	)
	return err
}

func (r *SecurityEventRepository) GetEventsByParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, limit int) ([]*models.SecurityEvent, error) {
	var events []*models.SecurityEvent
	err := r.db.SelectContext(ctx, &events, GET_EVENTS_BY_PARSER, parserID, orgID, types.ParseStatusSuccess, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *SecurityEventRepository) GetEventVolume(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time, interval string, sourceID *uuid.UUID) ([]map[string]interface{}, error) {
	trunc := "hour"
	switch interval {
	case "minute":
		trunc = "minute"
	case "day":
		trunc = "day"
	}
	query := fmt.Sprintf(`
		SELECT date_trunc('%s', occurred_at) AS timestamp,
		       COUNT(*) AS event_count,
		       COALESCE(source_id, '00000000-0000-0000-0000-000000000000'::uuid) AS source_id
		FROM security_events
		WHERE organization_id = $1
	`, trunc)
	args := []interface{}{orgID}
	paramIdx := 2

	if sourceID != nil {
		query += fmt.Sprintf(" AND source_id = $%d", paramIdx)
		args = append(args, *sourceID)
		paramIdx++
	}
	if startTime != nil && !startTime.IsZero() {
		query += fmt.Sprintf(" AND occurred_at >= $%d", paramIdx)
		args = append(args, *startTime)
		paramIdx++
	}
	if endTime != nil && !endTime.IsZero() {
		query += fmt.Sprintf(" AND occurred_at <= $%d", paramIdx)
		args = append(args, *endTime)
	}
	query += " GROUP BY date_trunc($1, occurred_at), source_id ORDER BY timestamp"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var ts time.Time
		var count int64
		var sid uuid.UUID
		err := rows.Scan(&ts, &count, &sid)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"timestamp":   ts,
			"source_id":   sid.String(),
			"event_count": count,
		})
	}
	return results, nil
}

func (r *SecurityEventRepository) GetEventCountInWindow(ctx context.Context, orgID uuid.UUID, startTime, endTime *time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM security_events WHERE organization_id = $1`
	args := []interface{}{orgID}
	if startTime != nil && !startTime.IsZero() {
		query += " AND occurred_at >= $2"
		args = append(args, *startTime)
	}
	if endTime != nil && !endTime.IsZero() {
		query += " AND occurred_at <= $3"
		args = append(args, *endTime)
	}
	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}
