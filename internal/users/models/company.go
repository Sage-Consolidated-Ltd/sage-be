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
	Role string `json:"role" db:"role"`
	Status string `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
type GetOrganizationResponse struct {
	ID   string `json:"id"`
Name string `json:"name"`
OwnerID string `json:"owner_id"`
Industry string `json:"industry"`
Role string `json:"role"`
Status string `json:"status"`
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
		Role: o.Role,
		Status: o.Status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

type OrganizationInvite struct {
	ID             string `json:"id" db:"id"`
	OrganizationID string `json:"organization_id" db:"organization_id"`
	Email          string `json:"email" db:"email"`
	RoleID         *string `json:"role_id" db:"role_id"`
	Status         string `json:"status" db:"status"`
	InvitedBy      *string `json:"invited_by" db:"invited_by"`
	TokenHash      *string `json:"token_hash" db:"token_hash"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
type OrganizationMember struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
type OrganizationRole struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
type GetOrganizationRolesResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
}
func (r *OrganizationRole) ToGetOrganizationRolesResponse() GetOrganizationRolesResponse {
	return GetOrganizationRolesResponse{
		ID: r.ID,
		Name: r.Name,
		Description: r.Description,
	}
}
