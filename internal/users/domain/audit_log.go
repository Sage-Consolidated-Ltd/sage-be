package domain

import (
	"database/sql"
	"time"
)

type AuditLog struct {
	ID             string         `json:"id" db:"id"`
	UserID         sql.NullString `json:"user_id" db:"user_id"`
	OrganizationID sql.NullString `json:"organization_id" db:"organization_id"`
	ActionType     string         `json:"action_type" db:"action_type"`
	ResourceType   string         `json:"resource_type" db:"resource_type"`
	ResourceID     sql.NullString `json:"resource_id" db:"resource_id"`
	Metadata       []byte         `json:"metadata" db:"metadata"`
	IPAddress      sql.NullString `json:"ip_address" db:"ip_address"`
	UserAgent      sql.NullString `json:"user_agent" db:"user_agent"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

type UserActivityResponse struct {
	ID           string                 `json:"id"`
	ActionType   string                 `json:"action_type"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

func (a *AuditLog) ToResponse() *UserActivityResponse {
	resp := &UserActivityResponse{
		ID:           a.ID,
		ActionType:   a.ActionType,
		ResourceType: a.ResourceType,
		CreatedAt:    a.CreatedAt,
	}
	if a.ResourceID.Valid {
		resp.ResourceID = a.ResourceID.String
	}
	return resp
}
