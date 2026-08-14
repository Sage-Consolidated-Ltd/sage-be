package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/outbound"

	"github.com/google/uuid"
)

type IncidentRepository struct {
	db *db.DB
}

func NewIncidentRepository(database *db.DB) outbound.IncidentRepository {
	return &IncidentRepository{db: database}
}

type incidentDB struct {
	ID             uuid.UUID             `db:"id"`
	OrganizationID uuid.UUID             `db:"organization_id"`
	RuleID         string                `db:"rule_id"`
	RuleName       string                `db:"rule_name"`
	Category       domain.RuleCategory   `db:"category"`
	Severity       string                `db:"severity"`
	Status         domain.IncidentStatus `db:"status"`
	Title          string                `db:"title"`
	Summary        string                `db:"summary"`
	Evidence       db.JSONMap            `db:"evidence"`
	OccurredAt     string                `db:"occurred_at"`
	CreatedAt      string                `db:"created_at"`
	UpdatedAt      string                `db:"updated_at"`
}

func (r *IncidentRepository) SaveIncident(ctx context.Context, incident *domain.Incident) error {
	if incident == nil {
		return fmt.Errorf("cannot save nil incident")
	}

	evidenceBytes, err := json.Marshal(incident.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal incident evidence: %w", err)
	}

	var evidenceMap db.JSONMap
	if err := json.Unmarshal(evidenceBytes, &evidenceMap); err != nil {
		evidenceMap = db.JSONMap{}
	}

	query := `
		INSERT INTO incidents (
			id, organization_id, rule_id, rule_name, category, severity,
			status, title, summary, evidence, occurred_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			evidence = EXCLUDED.evidence,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(
		ctx, query,
		incident.ID, incident.OrganizationID, incident.RuleID, incident.RuleName,
		incident.Category, string(incident.Severity), incident.Status, incident.Title,
		incident.Summary, evidenceMap, incident.OccurredAt, incident.CreatedAt, incident.UpdatedAt,
	)

	return err
}

func (r *IncidentRepository) GetIncidentByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Incident, error) {
	query := `
		SELECT id, organization_id, rule_id, rule_name, category, severity, status, title, summary, evidence, occurred_at, created_at, updated_at
		FROM incidents
		WHERE id = $1 AND organization_id = $2
	`

	var row incidentDB
	if err := r.db.GetContext(ctx, &row, query, id, orgID); err != nil {
		return nil, err
	}

	return toDomainIncident(&row)
}

func (r *IncidentRepository) ListIncidents(ctx context.Context, orgID uuid.UUID, status *domain.IncidentStatus, limit, offset int) ([]*domain.Incident, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM incidents WHERE organization_id = $1`
	query := `
		SELECT id, organization_id, rule_id, rule_name, category, severity, status, title, summary, evidence, occurred_at, created_at, updated_at
		FROM incidents
		WHERE organization_id = $1
	`

	args := []interface{}{orgID}
	if status != nil && *status != "" {
		countQuery += ` AND status = $2`
		query += ` AND status = $2`
		args = append(args, *status)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		total = 0
	}

	query += ` ORDER BY occurred_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	var rows []incidentDB
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Incident, 0, len(rows))
	for _, row := range rows {
		inc, err := toDomainIncident(&row)
		if err == nil {
			result = append(result, inc)
		}
	}

	return result, total, nil
}

func (r *IncidentRepository) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status domain.IncidentStatus) error {
	query := `UPDATE incidents SET status = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3`
	_, err := r.db.ExecContext(ctx, query, status, id, orgID)
	return err
}

func toDomainIncident(row *incidentDB) (*domain.Incident, error) {
	var evidence domain.Evidence
	if row.Evidence != nil {
		b, err := json.Marshal(row.Evidence)
		if err == nil {
			_ = json.Unmarshal(b, &evidence)
		}
	}

	return &domain.Incident{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		RuleID:         row.RuleID,
		RuleName:       row.RuleName,
		Category:       row.Category,
		Severity:       types.Severity(row.Severity),
		Status:         row.Status,
		Title:          row.Title,
		Summary:        row.Summary,
		Evidence:       evidence,
	}, nil
}
