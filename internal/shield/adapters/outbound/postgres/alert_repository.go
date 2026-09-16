package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type AlertRepository struct {
	db *db.DB
}

func NewAlertRepository(database *db.DB) outbound.AlertRepository {
	return &AlertRepository{db: database}
}

type alertDB struct {
	ID              uuid.UUID  `db:"id"`
	OrganizationID  uuid.UUID  `db:"organization_id"`
	ThreatLabel     string     `db:"threat_label"`
	MITRE           string     `db:"mitre"`
	LogSource       string     `db:"log_source"`
	EventID         string     `db:"event_id"`
	EntityHost      *string    `db:"entity_host"`
	EntityAccount   *string    `db:"entity_account"`
	EntityIP        *string    `db:"entity_ip"`
	SecurityEventID *uuid.UUID `db:"security_event_id"`
	Context         db.JSONMap `db:"context"`
	DetectedAt      time.Time  `db:"detected_at"`
	CreatedAt       time.Time  `db:"created_at"`
}

func (r *AlertRepository) SaveAlert(ctx context.Context, alert *domain.Alert) error {
	if alert == nil {
		return fmt.Errorf("cannot save nil alert")
	}
	return r.BulkSaveAlerts(ctx, []*domain.Alert{alert})
}

func (r *AlertRepository) BulkSaveAlerts(ctx context.Context, alerts []*domain.Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	chunkSize := 100
	for i := 0; i < len(alerts); i += chunkSize {
		end := i + chunkSize
		if end > len(alerts) {
			end = len(alerts)
		}
		chunk := alerts[i:end]

		if err := r.insertAlertChunk(ctx, chunk); err != nil {
			return err
		}
	}

	return nil
}

func (r *AlertRepository) insertAlertChunk(ctx context.Context, chunk []*domain.Alert) error {
	var valueStrings []string
	var valueArgs []interface{}
	paramIdx := 1

	for _, a := range chunk {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		detectedAt := a.DetectedAt
		if detectedAt.IsZero() {
			detectedAt = time.Now()
		}

		ctxBytes, _ := json.Marshal(a.Context)
		var ctxMap db.JSONMap
		if err := json.Unmarshal(ctxBytes, &ctxMap); err != nil {
			ctxMap = db.JSONMap{}
		}

		var secEvtID *uuid.UUID
		if a.RawEvent != nil && a.RawEvent.ID != uuid.Nil {
			secEvtID = &a.RawEvent.ID
		}

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			paramIdx, paramIdx+1, paramIdx+2, paramIdx+3, paramIdx+4,
			paramIdx+5, paramIdx+6, paramIdx+7, paramIdx+8, paramIdx+9,
			paramIdx+10, paramIdx+11, paramIdx+12,
		))

		valueArgs = append(valueArgs,
			a.ID,
			a.OrganizationID,
			a.ThreatLabel,
			a.MITRE,
			a.LogSource,
			a.EventID,
			stringToPtr(a.EntityHost),
			stringToPtr(a.EntityAccount),
			stringToPtr(a.EntityIP),
			secEvtID,
			ctxMap,
			detectedAt,
			time.Now(),
		)
		paramIdx += 13
	}

	query := fmt.Sprintf(`
		INSERT INTO alerts (
			id, organization_id, threat_label, mitre, log_source, event_id,
			entity_host, entity_account, entity_ip, security_event_id, context,
			detected_at, created_at
		) VALUES %s
		ON CONFLICT (id) DO NOTHING
	`, strings.Join(valueStrings, ", "))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	return err
}

func (r *AlertRepository) GetAlertByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, organization_id, threat_label, mitre, log_source, event_id,
		       entity_host, entity_account, entity_ip, security_event_id, context,
		       detected_at, created_at
		FROM alerts
		WHERE id = $1 AND organization_id = $2
	`

	var row alertDB
	if err := r.db.GetContext(ctx, &row, query, id, orgID); err != nil {
		return nil, err
	}

	return toDomainAlert(&row), nil
}

func (r *AlertRepository) ListAlerts(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]*domain.Alert, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM alerts WHERE organization_id = $1`
	query := `
		SELECT id, organization_id, threat_label, mitre, log_source, event_id,
		       entity_host, entity_account, entity_ip, security_event_id, context,
		       detected_at, created_at
		FROM alerts
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	paramIdx := 2

	if label, ok := filters["threat_label"].(string); ok && label != "" {
		countQuery += fmt.Sprintf(" AND threat_label = $%d", paramIdx)
		query += fmt.Sprintf(" AND threat_label = $%d", paramIdx)
		args = append(args, label)
		paramIdx++
	}
	if host, ok := filters["entity_host"].(string); ok && host != "" {
		countQuery += fmt.Sprintf(" AND entity_host = $%d", paramIdx)
		query += fmt.Sprintf(" AND entity_host = $%d", paramIdx)
		args = append(args, host)
		paramIdx++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		total = 0
	}

	query += fmt.Sprintf(" ORDER BY detected_at DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, limit, offset)

	var rows []alertDB
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Alert, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainAlert(&row))
	}

	return result, total, nil
}

func toDomainAlert(row *alertDB) *domain.Alert {
	var host, account, ip string
	if row.EntityHost != nil {
		host = *row.EntityHost
	}
	if row.EntityAccount != nil {
		account = *row.EntityAccount
	}
	if row.EntityIP != nil {
		ip = *row.EntityIP
	}

	ctxMap := make(map[string]interface{})
	if row.Context != nil {
		for k, v := range row.Context {
			ctxMap[k] = v
		}
	}

	return &domain.Alert{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		ThreatLabel:    row.ThreatLabel,
		MITRE:          row.MITRE,
		LogSource:      row.LogSource,
		EventID:        row.EventID,
		EntityHost:     host,
		EntityAccount:  account,
		EntityIP:       ip,
		Context:        ctxMap,
		DetectedAt:     row.DetectedAt,
	}
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

