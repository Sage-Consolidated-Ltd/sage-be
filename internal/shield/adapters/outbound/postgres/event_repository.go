package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/adapters/outbound/postgres/models"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type SecurityEventRepository struct {
	db *db.DB
}

func NewSecurityEventRepository(db *db.DB) outbound.SecurityEventRepository {
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
		ON CONFLICT (source, source_event_id) 
		DO NOTHING
	`
	BULK_INSERT_EVENTS_WITH_RETURNING = `
		WITH new_events AS (
			INSERT INTO security_events (
				organization_id, source_id, parser_id, source_event_id, source,
				event_type, event_category, severity, actor_email, actor_username,
				ip_address, geo_country, geo_city, raw_payload, normalized_payload,
				parse_status, parse_errors, occurred_at, ingested_at
			) VALUES %s
			RETURNING id, created_at
		)
		SELECT id FROM new_events
	`
	GET_EVENT     = `SELECT * FROM security_events WHERE id = $1 AND organization_id = $2`
	SEARCH_EVENTS = `
		SELECT * FROM security_events
		WHERE organization_id = $1
			AND ($2::uuid IS NULL OR source_id = $2)
			AND ($3::varchar IS NULL OR source = $3)
			AND ($4::varchar IS NULL OR event_type = $4)
			AND ($5::varchar IS NULL OR event_category = $5)
			AND (
				$6::varchar IS NULL
				OR severity = $6
				OR severity IS NULL
			)
			AND ($7::varchar IS NULL OR actor_email ILIKE '%' || $7 || '%')
			AND ($8::varchar IS NULL OR ip_address ILIKE '%' || $8 || '%')
			AND ($9::timestamptz IS NULL OR occurred_at >= $9)
			AND ($10::timestamptz IS NULL OR occurred_at <= $10)
			AND ($11::varchar IS NULL OR $11 = '' OR search_vector @@ plainto_tsquery('english', $11))
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
			AND ($11::varchar IS NULL OR $11 = '' OR search_vector @@ plainto_tsquery('english', $11))
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
	GET_RAW_EVENT_BY_ID  = `
		WITH picked AS (
			SELECT id
			FROM raw_events
			WHERE 
				id = $1
				AND organization_id = $2
				AND (
					locked_at IS NULL 
					OR locked_at < NOW() - INTERVAL '5 minutes'
				)
			FOR UPDATE SKIP LOCKED
		)
		UPDATE raw_events re
		SET 
			locked_at = NOW(),
			locked_by = $3
		FROM picked
		WHERE re.id = picked.id
		RETURNING re.*;`
)

func (r *SecurityEventRepository) CreateEvent(ctx context.Context, event *domain.SecurityEvent) error {
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
		parseErrs = make([]db.JSONMap, 0)
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

func (r *SecurityEventRepository) BulkCreateEvents(ctx context.Context, events []*domain.SecurityEvent) error {
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
			parseErrs = make([]db.JSONMap, 0)
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

func (r *SecurityEventRepository) GetEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.SecurityEvent, error) {
	var dto models.SecurityEventDTO
	err := r.db.GetContext(ctx, &dto, GET_EVENT, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("EVENT NOT FOUND")
		}
		return nil, err
	}
	return dto.ToDomain(), nil
}

func (r *SecurityEventRepository) SearchEvents(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	var sourceID *uuid.UUID
	if sid, ok := filters["source_id"].(uuid.UUID); ok && sid != uuid.Nil {
		sourceID = &sid
	} else if sidStr, ok := filters["source_id"].(string); ok && sidStr != "" {
		if uid, err := uuid.Parse(sidStr); err == nil {
			sourceID = &uid
		}
	}

	var source *string
	if s, ok := filters["source"].(string); ok && s != "" {
		source = &s
	}

	var eventType *string
	if et, ok := filters["event_type"].(string); ok && et != "" {
		eventType = &et
	}

	var category *string
	if c, ok := filters["event_category"].(string); ok && c != "" {
		category = &c
	} else if c, ok := filters["category"].(string); ok && c != "" {
		category = &c
	}

	var severity *string
	if s, ok := filters["severity"].(string); ok && s != "" {
		severity = &s
	}

	var actorEmail *string
	if ae, ok := filters["actor_email"].(string); ok && ae != "" {
		actorEmail = &ae
	}

	var ipAddress *string
	if ip, ok := filters["ip_address"].(string); ok && ip != "" {
		ipAddress = &ip
	}

	var startTime *time.Time
	if st, ok := filters["start_time"].(time.Time); ok && !st.IsZero() {
		startTime = &st
	}

	var endTime *time.Time
	if et, ok := filters["end_time"].(time.Time); ok && !et.IsZero() {
		endTime = &et
	}

	var search *string
	if q, ok := filters["q"].(string); ok && q != "" {
		search = &q
	} else if s, ok := filters["search"].(string); ok && s != "" {
		search = &s
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_EVENTS,
		orgID, sourceID, source, eventType, category, severity, actorEmail, ipAddress, startTime, endTime, search)
	if err != nil {
		return nil, 0, err
	}

	var dtos []*models.SecurityEventDTO
	err = r.db.SelectContext(ctx, &dtos, SEARCH_EVENTS,
		orgID, sourceID, source, eventType, category, severity, actorEmail, ipAddress,
		startTime, endTime, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.SecurityEvent, 0, len(dtos))
	for _, dto := range dtos {
		events = append(events, dto.ToDomain())
	}
	return events, total, nil
}

func (r *SecurityEventRepository) GetEventsBySource(ctx context.Context, sourceID uuid.UUID, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SecurityEvent, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	var eventType *string
	if et, ok := filters["event_type"].(string); ok && et != "" {
		eventType = &et
	}

	var severity *string
	if s, ok := filters["severity"].(string); ok && s != "" {
		severity = &s
	}

	var startTime *time.Time
	if st, ok := filters["start_time"].(time.Time); ok && !st.IsZero() {
		startTime = &st
	}

	var endTime *time.Time
	if et, ok := filters["end_time"].(time.Time); ok && !et.IsZero() {
		endTime = &et
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_EVENTS_BY_SOURCE,
		orgID, sourceID, eventType, severity, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}

	var dtos []*models.SecurityEventDTO
	err = r.db.SelectContext(ctx, &dtos, GET_EVENTS_BY_SOURCE,
		orgID, sourceID, eventType, severity, startTime, endTime, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.SecurityEvent, 0, len(dtos))
	for _, dto := range dtos {
		events = append(events, dto.ToDomain())
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

func (r *SecurityEventRepository) GetEventsByParser(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID, limit int) ([]*domain.SecurityEvent, error) {
	var dtos []*models.SecurityEventDTO
	err := r.db.SelectContext(ctx, &dtos, GET_EVENTS_BY_PARSER, parserID, orgID, types.ParseStatusSuccess, limit)
	if err != nil {
		return nil, err
	}

	events := make([]*domain.SecurityEvent, 0, len(dtos))
	for _, dto := range dtos {
		events = append(events, dto.ToDomain())
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

func (r *SecurityEventRepository) BulkCreateEventsWithReturning(ctx context.Context, events []*domain.SecurityEvent) ([]uuid.UUID, error) {
	if len(events) == 0 {
		return []uuid.UUID{}, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const eventInsertColumnCount = 19

	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*eventInsertColumnCount)

	for i, event := range events {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*eventInsertColumnCount+1,
				i*eventInsertColumnCount+2,
				i*eventInsertColumnCount+3,
				i*eventInsertColumnCount+4,
				i*eventInsertColumnCount+5,
				i*eventInsertColumnCount+6,
				i*eventInsertColumnCount+7,
				i*eventInsertColumnCount+8,
				i*eventInsertColumnCount+9,
				i*eventInsertColumnCount+10,
				i*eventInsertColumnCount+11,
				i*eventInsertColumnCount+12,
				i*eventInsertColumnCount+13,
				i*eventInsertColumnCount+14,
				i*eventInsertColumnCount+15,
				i*eventInsertColumnCount+16,
				i*eventInsertColumnCount+17,
				i*eventInsertColumnCount+18,
				i*eventInsertColumnCount+19,
			),
		)

		raw := event.RawPayload
		if raw == nil {
			raw = map[string]interface{}{}
		}

		rawJSON, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf(
				"marshal raw payload for event %v: %w",
				event.SourceEventID,
				err,
			)
		}

		norm := event.NormalizedPayload
		if norm == nil {
			norm = map[string]interface{}{}
		}

		normJSON, err := json.Marshal(norm)
		if err != nil {
			return nil, fmt.Errorf(
				"marshal normalized payload for event %v: %w",
				event.SourceEventID,
				err,
			)
		}

		parseErrs := event.ParseErrors
		if parseErrs == nil {
			parseErrs = []db.JSONMap{}
		}

		parseErrsJSON, err := json.Marshal(parseErrs)
		if err != nil {
			return nil, fmt.Errorf(
				"marshal parse errors for event %v: %w",
				event.SourceEventID,
				err,
			)
		}

		valueArgs = append(valueArgs,
			event.OrganizationID,
			event.SourceID,
			event.ParserID,
			event.SourceEventID,
			event.Source,

			event.EventType,
			event.EventCategory,
			event.Severity,
			event.ActorEmail,
			event.ActorUsername,

			event.IPAddress,
			event.GeoCountry,
			event.GeoCity,

			rawJSON,
			normJSON,

			event.ParseStatus,
			parseErrsJSON,

			event.OccurredAt,

			time.Now().UTC(), // ingested_at
		)
	}

	query := fmt.Sprintf(BULK_INSERT_EVENTS_WITH_RETURNING, strings.Join(valueStrings, ","))
	rows, err := tx.QueryContext(ctx, query, valueArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *SecurityEventRepository) BulkInsertRawEvents(
	ctx context.Context,
	orgID uuid.UUID,
	sourceID *uuid.UUID,
	events []domain.NormalizedEvent,
) ([]domain.CreateRawEventResponse, error) {

	if len(events) == 0 {
		return nil, nil
	}

	var (
		values       []string
		args         []interface{}
		placeholderN = 1
	)

	for _, event := range events {
		rawPayload, err := json.Marshal(event.Raw)
		if err != nil {
			return nil, err
		}

		provider_status := event.Status
		if provider_status == "" {
			provider_status = "pending"
		}

		values = append(values,
			fmt.Sprintf(
				`($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)`,
				placeholderN,
				placeholderN+1,
				placeholderN+2,
				placeholderN+3,
				placeholderN+4,
				placeholderN+5,
				placeholderN+6,
				placeholderN+7,
				placeholderN+8,
				placeholderN+9,
				placeholderN+10,
				placeholderN+11,
			),
		)

		args = append(args,
			orgID,
			sourceID,
			event.Provider,
			event.EventType,
			nullIfEmpty(event.UserID),
			nullIfEmpty(event.UserName),
			nullIfEmpty(event.IPAddress),
			nullIfEmpty(event.Application),
			event.Timestamp,
			provider_status,
			rawPayload,
			"pending",
		)

		placeholderN += 12
	}

	query := fmt.Sprintf(`
		INSERT INTO raw_events (
			organization_id,
			source_id,
			provider,
			event_type,
			user_id,
			user_name,
			ip_address,
			application,
			event_timestamp,
			provider_status,
			raw_payload,
			status
		)
		VALUES %s
		RETURNING id, collected_at
	`, strings.Join(values, ","))

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.CreateRawEventResponse

	for rows.Next() {
		var r domain.CreateRawEventResponse
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, rows.Err()
}

func (r *SecurityEventRepository) GetRawEventByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.RawEvent, error) {
	var dto models.RawEventDTO
	err := r.db.GetContext(ctx, &dto, GET_RAW_EVENT_BY_ID, id, orgID, "worker-1")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("RAW EVENT NOT FOUND")
		}
		return nil, err
	}
	return dto.ToDomain(), nil
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func (r *SecurityEventRepository) GetThreatsSummary(ctx context.Context, orgID uuid.UUID) (*domain.ThreatsSummary, error) {
	const q = `
		SELECT
			COALESCE(SUM(critical), 0) AS critical,
			COALESCE(SUM(high), 0) AS high,
			COALESCE(SUM(medium), 0) AS medium,
			COALESCE(SUM(low), 0) AS low,
			COALESCE(SUM(new_in_last_7_days), 0) AS new_in_last_7_days,
			COALESCE(SUM(total_threats), 0) AS total_threats
		FROM (
			SELECT
				COUNT(*) FILTER (WHERE LOWER(severity) = 'critical') AS critical,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'high') AS high,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'medium') AS medium,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'low') AS low,
				COUNT(*) FILTER (WHERE occurred_at >= NOW() - INTERVAL '7 days') AS new_in_last_7_days,
				COUNT(*) AS total_threats
			FROM security_events
			WHERE organization_id = $1
			UNION ALL
			SELECT
				COUNT(*) FILTER (WHERE LOWER(severity) = 'critical') AS critical,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'high') AS high,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'medium') AS medium,
				COUNT(*) FILTER (WHERE LOWER(severity) = 'low') AS low,
				COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days') AS new_in_last_7_days,
				COUNT(*) AS total_threats
			FROM threats
			WHERE organization_id = $1
		) combined
	`
	var dto models.ThreatsSummaryDTO
	if err := r.db.GetContext(ctx, &dto, q, orgID); err != nil {
		return nil, fmt.Errorf("failed to get threats summary: %w", err)
	}
	return dto.ToDomain(), nil
}
