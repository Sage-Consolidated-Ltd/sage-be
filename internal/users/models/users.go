package models

import (
	"database/sql"
	"sage-backend/internal/shared/types"
	"time"
)

type User struct {
	ID           string       `json:"id" db:"id"`
	FirstName    string       `json:"first_name" db:"first_name"`
	LastName     string       `json:"last_name" db:"last_name"`
	Email        string       `json:"email" db:"email"`
	IsVerified   bool         `json:"is_verified" db:"is_verified"`
	PasswordHash string       `json:"password_hash" db:"password_hash"`
	Role         types.Role   `json:"role" db:"role"`
	TwoFactorSecret sql.NullString `json:"two_factor_secret" db:"two_factor_secret"`
	TwoFactorEnabled bool `json:"two_factor_enabled" db:"two_factor_enabled"`
	TimeZone sql.NullString `json:"time_zone" db:"time_zone"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

type ExternalUser struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type GetUserResponse struct {
	ID string `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TwoFactorEnabled bool `json:"two_factor_enabled" db:"two_factor_enabled"`
	CreatedAt time.Time `json:"created_at"`
	Organization []GetOrganizationResponse `json:"organization,omitempty"`
	TimeZone *string `json:"time_zone,omitempty"`
}
func(u *User) ToResponse(orgs *[]Organization) *GetUserResponse {
	var organizationsResp []GetOrganizationResponse
	if orgs != nil {
		for _, org := range *orgs {
			organizationsResp = append(organizationsResp, *org.ToResponse())
		}
	}
	return &GetUserResponse{
		ID: u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      string(u.Role),
		TwoFactorEnabled: u.TwoFactorEnabled,
		TimeZone: &u.TimeZone.String,
		CreatedAt: u.CreatedAt,
		Organization: organizationsResp,
	}
}
