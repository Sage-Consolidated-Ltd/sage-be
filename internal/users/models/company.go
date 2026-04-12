package models

import (
	"database/sql"
	"time"
)

type Industry struct {
	ID        string     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}
type GetIndustriesResponse struct {
	ID  string `json:"id"`
	Name string `json:"name"`
}
func (i *Industry) ToGetIndustriesResponse() GetIndustriesResponse {
	return GetIndustriesResponse{
		ID: i.ID,
		Name: i.Name,
	}
}

type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OwnerID   string    `json:"owner_id" db:"owner_id"`
	Industry string `json:"industry" db:"industry"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
type GetOrganizationResponse struct {
	ID   string `json:"id"`
Name string `json:"name"`
OwnerID string `json:"owner_id"`
Industry string `json:"industry"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
func (o *Organization) ToResponse() *GetOrganizationResponse {
	var deletedAt *time.Time
	if o.DeletedAt.Valid {
		deletedAt = &o.DeletedAt.Time
	}
	return &GetOrganizationResponse{
		ID: o.ID,
		Name: o.Name,
		OwnerID: o.OwnerID,
		Industry: o.Industry,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
