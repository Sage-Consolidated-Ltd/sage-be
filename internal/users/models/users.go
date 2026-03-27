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
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
