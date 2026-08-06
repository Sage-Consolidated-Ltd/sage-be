package postgres

import (
	"database/sql"
	"sage-backend/internal/organization/domain"
	"time"
)

type organizationModel struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	Slug                 string         `db:"slug"`
	OwnerID              string         `db:"owner_id"`
	Industry             string         `db:"industry"`
	Country              sql.NullString `db:"country"`
	Timezone             string         `db:"timezone"`
	RiskThresholdDefault int            `db:"risk_threshold_default"`
	Role                 string         `db:"role"`
	Status               string         `db:"status"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	DeletedAt            sql.NullTime   `db:"deleted_at"`
}

func (m organizationModel) ToDomain() domain.Organization {
	return domain.Organization{
		ID:                   m.ID,
		Name:                 m.Name,
		Slug:                 m.Slug,
		OwnerID:              m.OwnerID,
		Industry:             m.Industry,
		Country:              m.Country,
		Timezone:             m.Timezone,
		RiskThresholdDefault: m.RiskThresholdDefault,
		Role:                 m.Role,
		Status:               m.Status,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		DeletedAt:            m.DeletedAt,
	}
}
