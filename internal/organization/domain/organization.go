package domain

import (
	"database/sql"
	"time"
)

type Industry struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
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
	PrimaryContactEmail  Email          `json:"primary_contact_email" db:"primary_contact_email"`
	SupportEmail         Email          `json:"support_email" db:"support_email"`
	RiskThresholdDefault int            `json:"risk_threshold_default" db:"risk_threshold_default"`
	Role                 string         `json:"role" db:"role"`
	Status               string         `json:"status" db:"status"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt            sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type OrganizationEnvironment struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	SubscriptionID string    `json:"subscription_id" db:"subscription_id"`
	Region         string    `json:"region" db:"region"`
	DeploymentMode string    `json:"deployment_mode" db:"deployment_mode"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationBranding struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	LogoLightURL   string    `json:"logo_light_url" db:"logo_light_url"`
	LogoDarkURL    string    `json:"logo_dark_url" db:"logo_dark_url"`
	ShowInReports  bool      `json:"show_in_reports" db:"show_in_reports"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationMember struct {
	ID             string         `json:"id" db:"id"`
	OrganizationID string         `json:"organization_id" db:"organization_id"`
	UserID         string         `json:"user_id" db:"user_id"`
	Role           string         `json:"role" db:"role"`
	RoleID         sql.NullString `json:"role_id" db:"role_id"`
	Status         string         `json:"status" db:"status"`
	InvitedBy      sql.NullString `json:"invited_by" db:"invited_by"`
	InvitedAt      sql.NullTime   `json:"invited_at,omitempty" db:"invited_at"`
	JoinedAt       sql.NullTime   `json:"joined_at,omitempty" db:"joined_at"`
	Department     sql.NullString `json:"department,omitempty" db:"department"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`

	// Joined fields
	Email           string         `json:"email,omitempty" db:"email"`
	FirstName       string         `json:"first_name,omitempty" db:"first_name"`
	LastName        string         `json:"last_name,omitempty" db:"last_name"`
	AvatarURL       sql.NullString `json:"avatar_url,omitempty" db:"avatar_url"`
	UserName        string         `json:"user_name,omitempty" db:"user_name"`
	UserEmail       string         `json:"user_email,omitempty" db:"user_email"`
	UserAvatar      sql.NullString `json:"user_avatar,omitempty" db:"user_avatar"`
	TwoFactorStatus bool           `json:"two_factor_status" db:"two_factor_enabled"`
	LastLoginAt     sql.NullTime   `json:"last_login_at,omitempty" db:"last_login_at"`
}

type OrganizationInvite struct {
	ID             string         `json:"id" db:"id"`
	OrganizationID string         `json:"organization_id" db:"organization_id"`
	Email          string         `json:"email" db:"email"`
	RoleID         string         `json:"role_id" db:"role_id"`
	Status         string         `json:"status" db:"status"`
	InvitedBy      string         `json:"invited_by" db:"invited_by"`
	TokenHash      *string        `json:"token_hash" db:"token_hash"`
	ExpiresAt      time.Time      `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type OrganizationSettings struct {
	ID                           string         `json:"id" db:"id"`
	OrganizationID               string         `json:"organization_id" db:"organization_id"`
	DefaultAlertSeverityThreshold string         `json:"default_alert_severity_threshold" db:"default_alert_severity_threshold"`
	AutoContainmentEnabled       bool           `json:"auto_containment_enabled" db:"auto_containment_enabled"`
	AutoContainmentThreshold     int            `json:"auto_containment_threshold" db:"auto_containment_threshold"`
	AllowedIPRanges              sql.NullString `json:"allowed_ip_ranges" db:"allowed_ip_ranges"`
	SessionTimeoutMinutes        int            `json:"session_timeout_minutes" db:"session_timeout_minutes"`
	AuditLoggingEnabled          bool           `json:"audit_logging_enabled" db:"audit_logging_enabled"`
	CreatedAt                    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time      `json:"updated_at" db:"updated_at"`
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

type CustomRole struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	IsSystemRole   bool      `json:"is_system_role" db:"is_system_role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`

	// Populated via joins
	Permissions      []string `json:"permissions,omitempty"`
	PermissionGroups []string `json:"permission_groups,omitempty"`
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

type GetOrganizationResponse struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	OwnerID              string     `json:"owner_id"`
	Industry             string     `json:"industry"`
	Country              string     `json:"country,omitempty"`
	Timezone             string     `json:"timezone"`
	PrimaryContactEmail  string     `json:"primary_contact_email,omitempty"`
	SupportEmail         string     `json:"support_email,omitempty"`
	RiskThresholdDefault int        `json:"risk_threshold_default"`
	Role                 string     `json:"role"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

func (o *Organization) ToResponse() *GetOrganizationResponse {
	resp := &GetOrganizationResponse{
		ID:                   o.ID,
		Name:                 o.Name,
		Slug:                 o.Slug,
		OwnerID:              o.OwnerID,
		Industry:             o.Industry,
		Timezone:             o.Timezone,
		PrimaryContactEmail:  o.PrimaryContactEmail.String(),
		SupportEmail:         o.SupportEmail.String(),
		RiskThresholdDefault: o.RiskThresholdDefault,
		Role:                 o.Role,
		Status:               o.Status,
		CreatedAt:            o.CreatedAt,
		UpdatedAt:            o.UpdatedAt,
	}
	if o.Country.Valid {
		resp.Country = o.Country.String
	}
	if o.DeletedAt.Valid {
		resp.DeletedAt = &o.DeletedAt.Time
	}
	return resp
}

type OrganizationMemberResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (m *OrganizationMember) ToResponse() *OrganizationMemberResponse {
	email := m.Email
	if email == "" {
		email = m.UserEmail
	}
	fullName := m.FirstName + " " + m.LastName
	if (m.FirstName == "" && m.LastName == "") && m.UserName != "" {
		fullName = m.UserName
	}
	resp := &OrganizationMemberResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		Email:     email,
		FullName:  fullName,
		Role:      m.Role,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
	if m.AvatarURL.Valid {
		resp.AvatarURL = m.AvatarURL.String
	} else if m.UserAvatar.Valid {
		resp.AvatarURL = m.UserAvatar.String
	}
	if m.LastLoginAt.Valid {
		resp.LastLoginAt = m.LastLoginAt.Time
	}
	return resp
}

type OrganizationSettingsResponse struct {
	DefaultAlertSeverityThreshold string   `json:"default_alert_severity_threshold"`
	AutoContainmentEnabled        bool     `json:"auto_containment_enabled"`
	AutoContainmentThreshold      int      `json:"auto_containment_threshold"`
	AllowedIPRanges               []string `json:"allowed_ip_ranges"`
	SessionTimeoutMinutes         int      `json:"session_timeout_minutes"`
	AuditLoggingEnabled           bool     `json:"audit_logging_enabled"`
}

type OrganizationProfileCompanyDetailsResponse struct {
	Name                string `json:"organization_name"`
	Industry            string `json:"industry"`
	Region              string `json:"region"`
	PrimaryContactEmail string `json:"primary_contact_email"`
	SupportEmail        string `json:"support_email,omitempty"`
}

type OrganizationProfileEnvironmentInfoResponse struct {
	OrganizationID string `json:"organization_id"`
	TenantID       string `json:"tenant_id"`
	SubscriptionID string `json:"subscription_id"`
	DeploymentMode string `json:"deployment_mode"`
}

type OrganizationProfileBrandingResponse struct {
	LogoLightURL  string `json:"logo_light_url,omitempty"`
	LogoDarkURL   string `json:"logo_dark_url,omitempty"`
	ShowInReports bool   `json:"show_in_reports"`
}

type OrganizationProfileResponse struct {
	CompanyDetails         OrganizationProfileCompanyDetailsResponse  `json:"company_details"`
	EnvironmentInformation OrganizationProfileEnvironmentInfoResponse `json:"environment_information"`
	BrandingAppearance     OrganizationProfileBrandingResponse        `json:"branding_appearance"`
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
	if s.AllowedIPRanges.Valid && s.AllowedIPRanges.String != "" {
		resp.AllowedIPRanges = []string{s.AllowedIPRanges.String}
	}
	return resp
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

func (r *CustomRole) ToResponse() *CustomRoleResponse {
	return &CustomRoleResponse{
		ID:               r.ID,
		Name:             r.Name,
		Description:      r.Description,
		IsSystemRole:     r.IsSystemRole,
		PermissionGroups: r.PermissionGroups,
		Permissions:      r.Permissions,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

type PermissionGroupResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Permissions []string `json:"permissions,omitempty"`
}
