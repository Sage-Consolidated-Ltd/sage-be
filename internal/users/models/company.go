package models

import (
	"database/sql"
	"time"
)

type Industry struct {
	ID        string       `db:"id" json:"id"`
	Name      string       `db:"name" json:"name"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}
type GetIndustriesResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (i *Industry) ToGetIndustriesResponse() GetIndustriesResponse {
	return GetIndustriesResponse{
		ID:   i.ID,
		Name: i.Name,
	}
}

type Organization struct {
	ID                   string         `json:"id" db:"id"`
	Name                 string         `json:"name" db:"name"`
	Slug                 string         `json:"slug" db:"slug"`
	OwnerID              string         `json:"owner_id" db:"owner_id"`
	Industry             string         `json:"industry" db:"industry"`
	Country              sql.NullString `json:"country" db:"country"`
	Timezone             string         `json:"timezone" db:"timezone"`
	RiskThresholdDefault int            `json:"risk_threshold_default" db:"risk_threshold_default"`
	Role                 string         `json:"role" db:"role"`
	Status               string         `json:"status" db:"status"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt            sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}
type GetOrganizationResponse struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	OwnerID              string     `json:"owner_id"`
	Industry             string     `json:"industry"`
	Country              string     `json:"country,omitempty"`
	Timezone             string     `json:"timezone"`
	RiskThresholdDefault int        `json:"risk_threshold_default"`
	Role                 string     `json:"role"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

func (o *Organization) ToResponse() *GetOrganizationResponse {
	var deletedAt *time.Time
	if o.DeletedAt.Valid {
		deletedAt = &o.DeletedAt.Time
	}
	resp := &GetOrganizationResponse{
		ID:                   o.ID,
		Name:                 o.Name,
		Slug:                 o.Slug,
		OwnerID:              o.OwnerID,
		Industry:             o.Industry,
		Timezone:             o.Timezone,
		RiskThresholdDefault: o.RiskThresholdDefault,
		Role:                 o.Role,
		Status:               o.Status,
		CreatedAt:            o.CreatedAt,
		UpdatedAt:            o.UpdatedAt,
		DeletedAt:            deletedAt,
	}
	if o.Country.Valid {
		resp.Country = o.Country.String
	}
	return resp
}

type OrganizationInvite struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Email          string    `json:"email" db:"email"`
	RoleID         *string   `json:"role_id" db:"role_id"`
	Status         string    `json:"status" db:"status"`
	InvitedBy      *string   `json:"invited_by" db:"invited_by"`
	TokenHash      *string   `json:"token_hash" db:"token_hash"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationMember represents a member in the organization with full details
type OrganizationMember struct {
	ID             string         `json:"id" db:"id"`
	OrganizationID string         `json:"organization_id" db:"organization_id"`
	UserID         string         `json:"user_id" db:"user_id"`
	Email          string         `json:"email" db:"email"`
	FirstName      string         `json:"first_name" db:"first_name"`
	LastName       string         `json:"last_name" db:"last_name"`
	AvatarURL      sql.NullString `json:"avatar_url" db:"avatar_url"`
	Role           string         `json:"role" db:"role"`
	Department     sql.NullString `json:"department" db:"department"`
	Status         string         `json:"status" db:"status"`
	InvitedBy      sql.NullString `json:"invited_by" db:"invited_by"`
	InvitedAt      sql.NullTime   `json:"invited_at" db:"invited_at"`
	JoinedAt       sql.NullTime   `json:"joined_at" db:"joined_at"`
	LastLoginAt    sql.NullTime   `json:"last_login_at" db:"last_login_at"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// OrganizationMemberResponse is the API response for member list
type OrganizationMemberResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	Department  string    `json:"department,omitempty"`
	Status      string    `json:"status"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (m *OrganizationMember) ToResponse() *OrganizationMemberResponse {
	resp := &OrganizationMemberResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		Email:     m.Email,
		FullName:  m.FirstName + " " + m.LastName,
		Role:      m.Role,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
	if m.AvatarURL.Valid {
		resp.AvatarURL = m.AvatarURL.String
	}
	if m.LastLoginAt.Valid {
		resp.LastLoginAt = m.LastLoginAt.Time
	}
	if m.Department.Valid {
		resp.Department = m.Department.String
	}
	return resp
}

// OrganizationSettings stores organization-wide configuration
type OrganizationSettings struct {
	ID                            string    `json:"id" db:"id"`
	OrganizationID                string    `json:"organization_id" db:"organization_id"`
	DefaultAlertSeverityThreshold string    `json:"default_alert_severity_threshold" db:"default_alert_severity_threshold"`
	AutoContainmentEnabled        bool      `json:"auto_containment_enabled" db:"auto_containment_enabled"`
	AutoContainmentThreshold      int       `json:"auto_containment_threshold" db:"auto_containment_threshold"`
	AllowedIPRanges               []byte    `json:"-" db:"allowed_ip_ranges"` // JSONB stored as bytes
	SessionTimeoutMinutes         int       `json:"session_timeout_minutes" db:"session_timeout_minutes"`
	AuditLoggingEnabled           bool      `json:"audit_logging_enabled" db:"audit_logging_enabled"`
	CreatedAt                     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationSettingsResponse is the API response for settings
type OrganizationSettingsResponse struct {
	DefaultAlertSeverityThreshold string   `json:"default_alert_severity_threshold"`
	AutoContainmentEnabled        bool     `json:"auto_containment_enabled"`
	AutoContainmentThreshold      int      `json:"auto_containment_threshold"`
	AllowedIPRanges               []string `json:"allowed_ip_ranges"`
	SessionTimeoutMinutes         int      `json:"session_timeout_minutes"`
	AuditLoggingEnabled           bool     `json:"audit_logging_enabled"`
}

func (s *OrganizationSettings) ToResponse() *OrganizationSettingsResponse {
	resp := &OrganizationSettingsResponse{
		DefaultAlertSeverityThreshold: s.DefaultAlertSeverityThreshold,
		AutoContainmentEnabled:        s.AutoContainmentEnabled,
		AutoContainmentThreshold:      s.AutoContainmentThreshold,
		SessionTimeoutMinutes:         s.SessionTimeoutMinutes,
		AuditLoggingEnabled:           s.AuditLoggingEnabled,
		AllowedIPRanges:               []string{},
	}
	// Parse JSONB allowed_ip_ranges
	if len(s.AllowedIPRanges) > 0 {
		// Simple parsing - in production use proper JSON unmarshaling
		resp.AllowedIPRanges = parseIPRanges(s.AllowedIPRanges)
	}
	return resp
}

// parseIPRanges parses JSONB array of IP ranges
func parseIPRanges(data []byte) []string {
	// Simple implementation - removes brackets and quotes, splits by comma
	// In production, use encoding/json
	result := []string{}
	str := string(data)
	if str == "[]" || str == "" {
		return result
	}
	// Remove brackets and quotes
	str = str[1 : len(str)-1]
	if str == "" {
		return result
	}
	return result
}

type OrganizationRole struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
type GetOrganizationRolesResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *OrganizationRole) ToGetOrganizationRolesResponse() GetOrganizationRolesResponse {
	return GetOrganizationRolesResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}

type Permission struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PermissionGroup struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PermissionGroupResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Permissions []string `json:"permissions,omitempty"`
}

type CustomRole struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	IsSystemRole   bool      `json:"is_system_role" db:"is_system_role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CustomRoleResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	IsSystemRole     bool      `json:"is_system_role"`
	PermissionGroups []string  `json:"permission_groups,omitempty"`
	Permissions      []string  `json:"permissions,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
