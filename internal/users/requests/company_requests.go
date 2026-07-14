package requests

// Legacy invitation requests
type InviteMemberRequest struct {
	Email  string `json:"email" validate:"email,required"`
	RoleId string `json:"role_id" validate:"required"`
}

type BulkInviteMembersRequest struct {
	Invites []InviteMemberRequest `json:"invites" validate:"required,dive"`
}

// UpdateOrganizationRequest updates organization metadata
type UpdateOrganizationRequest struct {
	Name                 string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Industry             string `json:"industry,omitempty"`
	Country              string `json:"country,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
	RiskThresholdDefault int    `json:"risk_threshold_default,omitempty" validate:"omitempty,min=0,max=100"`
}

// ListMembersRequest query params for member list
type ListMembersRequest struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Role     string `query:"role,omitempty" validate:"omitempty,oneof=owner admin analyst viewer automation_admin billing_admin"`
	Search   string `query:"search,omitempty"`
}

// MemberInvite represents a single member invitation
type MemberInvite struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=admin analyst viewer automation_admin billing_admin"`
	ForceMFA bool   `json:"force_mfa"`
	Message  string `json:"message,omitempty" validate:"omitempty,max=500"`
}

// InviteMembersRequest invites users to join the organization
type InviteMembersRequest struct {
	Invites []MemberInvite `json:"invites" validate:"required,min=1,dive"`
}

// UpdateMemberRoleRequest updates a member's role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin analyst viewer automation_admin billing_admin"`
}

// UpdateOrganizationSettingsRequest updates org-level configuration
type UpdateOrganizationSettingsRequest struct {
	DefaultAlertSeverityThreshold string   `json:"default_alert_severity_threshold,omitempty" validate:"omitempty,oneof=low medium high critical"`
	AutoContainmentEnabled        bool     `json:"auto_containment_enabled"`
	AutoContainmentThreshold      int      `json:"auto_containment_threshold,omitempty" validate:"omitempty,min=0,max=100"`
	AllowedIPRanges               []string `json:"allowed_ip_ranges,omitempty"`
	SessionTimeoutMinutes         int      `json:"session_timeout_minutes,omitempty" validate:"omitempty,min=5,max=1440"`
	AuditLoggingEnabled           bool     `json:"audit_logging_enabled"`
}

// CreateCustomRoleRequest creates a custom role
type CreateCustomRoleRequest struct {
	Name             string   `json:"name" validate:"required,min=2,max=100"`
	Description      string   `json:"description,omitempty" validate:"omitempty,max=500"`
	PermissionGroups []string `json:"permission_groups,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
}

// UpdateCustomRoleRequest updates a custom role
type UpdateCustomRoleRequest struct {
	Name             string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description      string   `json:"description,omitempty" validate:"omitempty,max=500"`
	PermissionGroups []string `json:"permission_groups,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
}

type UpdateMemberStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended"`
}
