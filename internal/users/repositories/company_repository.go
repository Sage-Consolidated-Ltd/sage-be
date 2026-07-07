package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/requests"
	"strings"
	"time"
)

type CompanyRepositoryInt interface {
	// Industries & Roles
	GetIndustries(ctx context.Context) (*[]models.Industry, error)
	GetOrganizationRoles(ctx context.Context) (*[]models.OrganizationRole, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*models.OrganizationRole, error)
	GetIndustryByID(ctx context.Context, id string) (*models.Industry, error)

	// Legacy Invitation Flow
	InviteMemberToOrganization(ctx context.Context, req *requests.InviteMemberRequest, organizationId string, owner_id string, tokenHash *string, expiresAt time.Time) (*string, error)
	GetByID(ctx context.Context, id string) (*models.OrganizationInvite, error)
	MarkAccepted(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	GetMemberByEmail(ctx context.Context, email string, organizationId string) (*models.OrganizationMember, error)
	GetOrganizationByOwnerID(ctx context.Context, ownerId string) (*models.Organization, error)
	AddMemberToOrganization(ctx context.Context, organizationId string, userId string, roleId string) error

	// Organization Metadata
	GetOrganizationByID(ctx context.Context, orgID string) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, orgID string, req *requests.UpdateOrganizationRequest) error

	// Members
	ListMembers(ctx context.Context, orgID string, page, pageSize int, role, search string) ([]models.OrganizationMember, int, error)
	GetMemberByID(ctx context.Context, memberID, orgID string) (*models.OrganizationMember, error)
	AddMember(ctx context.Context, orgID, userID, role, invitedBy string) (*models.OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, memberID, orgID, newRole string) error
	UpdateMemberStatus(ctx context.Context, memberID, orgID, newStatus string) error
	ResetMember2FA(ctx context.Context, memberID, orgID string) error
	RemoveMember(ctx context.Context, memberID, orgID string) error
	GetMemberCountByRole(ctx context.Context, orgID, role string) (int, error)

	// Settings
	GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error)
	CreateOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error)
	UpdateOrganizationSettings(ctx context.Context, orgID string, req *requests.UpdateOrganizationSettingsRequest) error

	// Role helpers
	GetValidRoles() []string
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	ListPermissionGroups(ctx context.Context) ([]models.PermissionGroup, error)
	ListCustomRoles(ctx context.Context, orgID string) ([]models.CustomRole, error)
	GetCustomRoleByID(ctx context.Context, roleID, orgID string) (*models.CustomRole, error)
	CreateCustomRole(ctx context.Context, orgID string, req *requests.CreateCustomRoleRequest) (*models.CustomRole, error)
	UpdateCustomRole(ctx context.Context, roleID, orgID string, req *requests.UpdateCustomRoleRequest) error
	DeleteCustomRole(ctx context.Context, roleID, orgID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)
	GetRolePermissionGroups(ctx context.Context, roleID string) ([]string, error)
	SetRolePermissionGroups(ctx context.Context, roleID string, groupIDs []string) error
	SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	CheckMemberInOrganization(ctx context.Context, userID, orgID string) (bool, error)
	IsValidRole(role string) bool
	CanManageMembers(actorRole string) bool
	CanUpdateSettings(actorRole string) bool
}

type CompanyRepository struct {
	db *db.DB
}

var (
	GET_INDUSTRIES = `
		SELECT id, name FROM industries ORDER BY name ASC`
	GET_INDUSTRY_BY_ID = `
		SELECT id, name FROM industries WHERE id = $1`
	GET_INVITATION_BY_ID = `
	SELECT id, organization_id, email, role_id, status, invited_by,
	       token_hash, expires_at, created_at, updated_at
	FROM organization_invites
	WHERE id = $1
	`
	GET_MEMBER_BY_EMAIL = `
		SELECT 
		    om.id,
		    om.organization_id,
		    om.user_id,
		    u.email,
		    u.first_name,
		    u.last_name,
		    COALESCE(u.avatar_url, '') as avatar_url,
		    omr.name as role,
		    om.status,
		    om.created_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		JOIN organization_roles omr ON om.role_id = omr.id
		WHERE u.email = $1
		AND om.organization_id = $2;`
	MARK_INVITATION_ACCEPTED = `
		UPDATE organization_invites
			SET status = 'accepted', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'`
	MARK_INVITATION_EXPIRED = `
		UPDATE organization_invites
		SET status = 'expired', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		AND expires_at < NOW()`
	GET_ORGANIZATION_ROLES = `
		SELECT * FROM organization_roles
	`
	GET_ORGANIZATION_ROLE_BY_ID = `
		SELECT * FROM organization_roles WHERE id = $1
	`
	GET_ORGANIZATION_BY_OWNER_ID = `
		SELECT 
			o.id, 
			o.name, 
			o.owner_id, 
			COALESCE(i.name, '') AS industry,
			COALESCE(r.name, '') AS role, 
			om.status,
			o.created_at, 
			o.updated_at, 
			o.deleted_at 
		FROM organizations o
		JOIN organization_members om ON o.id = om.organization_id
		LEFT JOIN industries i ON o.industry_id = i.id
		LEFT JOIN organization_roles r ON om.role_id = r.id
		WHERE o.owner_id = $1
	`

	// Organization queries
	getOrganizationByIDSQL = `
		SELECT id, name, slug, owner_id, industry, country, timezone, 
		       risk_threshold_default, created_at, updated_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`
	updateOrganizationSQL = `
		UPDATE organizations
		SET name = COALESCE(NULLIF($2, ''), name),
		    industry = COALESCE(NULLIF($3, ''), industry),
		    country = COALESCE(NULLIF($4, ''), country),
		    timezone = COALESCE(NULLIF($5, ''), timezone),
		    risk_threshold_default = COALESCE(NULLIF($6, 0), risk_threshold_default),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`
	listMembersSQL = `
		SELECT 
		    om.id,
		    om.organization_id,
		    om.user_id,
		    u.email,
		    u.first_name,
		    u.last_name,
		    COALESCE(u.avatar_url, '') as avatar_url,
		    omr.name::text as role,
		    om.status::text,
		    om.invited_by,
		    om.invited_at,
		    om.joined_at,
		    u.last_login_at,
		    om.created_at,
		    om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		JOIN organization_roles omr ON om.role_id = omr.id
		WHERE om.organization_id = $1
		  AND ($2 = '' OR omr.name::text = $2)
		  AND ($3 = '' OR u.email ILIKE '%' || $3 || '%' OR u.first_name ILIKE '%' || $3 || '%' OR u.last_name ILIKE '%' || $3 || '%')
		  AND om.status != 'suspended'
		ORDER BY om.created_at DESC
		LIMIT $4 OFFSET $5
	`
	countMembersSQL = `
		SELECT COUNT(*)
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		JOIN organization_roles omr ON om.role_id = omr.id
		WHERE om.organization_id = $1
		  AND ($2 = '' OR omr.name::text = $2)
		  AND ($3 = '' OR u.email ILIKE '%' || $3 || '%' OR u.first_name ILIKE '%' || $3 || '%' OR u.last_name ILIKE '%' || $3 || '%')
		  AND om.status != 'suspended'
	`
	getMemberByIDSQL = `
		SELECT 
		    om.id,
		    om.organization_id,
		    om.user_id,
		    u.email,
		    u.first_name,
		    u.last_name,
		    COALESCE(u.avatar_url, '') as avatar_url,
		    omr.name::text as role,
		    om.status::text,
		    om.invited_by,
		    om.invited_at,
		    om.joined_at,
		    u.last_login_at,
		    om.created_at,
		    om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		JOIN organization_roles omr ON om.role_id = omr.id
		WHERE om.id = $1 AND om.organization_id = $2
	`
	getMemberByUserID = `
		SELECT
			om.id,
			om.organization_id,
			om.user_id,
			u.email,
			u.first_name,
			u.last_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			omr.name::text as role,
		    om.status::text,
		    om.invited_by,
		    om.invited_at,
		    om.joined_at,
		    u.last_login_at,
		    om.created_at,
		    om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		JOIN organization_roles omr ON om.role_id = omr.id
		WHERE u.id = $1 AND om.organization_id = $2
	`
	addMemberSQL = `
		INSERT INTO organization_members (organization_id, user_id, role, status, invited_by, invited_at)
		VALUES ($1, $2, $3::organization_role, 'invited', $4, NOW())
		RETURNING id, organization_id, user_id, role::text as role, status::text, invited_by, invited_at, created_at, updated_at
	`
	updateMemberRoleSQL = `
		UPDATE organization_members
		SET role = $3::organization_role,
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id
	`
	updateMemberStatusSQL = `
		UPDATE organization_members
		SET status = $3::member_status,
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id
	`
	checkMemberInOrganizationSQL = `
		SELECT EXISTS (
			SELECT 1
			FROM organization_members
			WHERE user_id = $1 AND organization_id = $2
		) AS is_member
	`
	resetMember2FASQL = `
		UPDATE users
		SET two_factor_enabled = false,
		    two_factor_secret = NULL,
		    updated_at = NOW()
		FROM organization_members om
		WHERE om.id = $1
		  AND om.organization_id = $2
		  AND users.id = om.user_id
		RETURNING users.id
	`
	removeMemberSQL = `
		UPDATE organization_members
		SET status = 'suspended',
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id
	`
	countMembersByRoleSQL = `
		SELECT COUNT(*)
		FROM organization_members
		WHERE organization_id = $1 AND role::text = $2 AND status != 'suspended'
	`
	getOrganizationSettingsSQL = `
		SELECT id, organization_id, default_alert_severity_threshold,
		       auto_containment_enabled, auto_containment_threshold,
		       allowed_ip_ranges, session_timeout_minutes, audit_logging_enabled,
		       created_at, updated_at
		FROM organization_settings
		WHERE organization_id = $1
	`
	createOrganizationSettingsSQL = `
		INSERT INTO organization_settings (organization_id)
		VALUES ($1)
		RETURNING id, organization_id, default_alert_severity_threshold,
		          auto_containment_enabled, auto_containment_threshold,
		          allowed_ip_ranges, session_timeout_minutes, audit_logging_enabled,
		          created_at, updated_at
	`
	updateOrganizationSettingsSQL = `
		UPDATE organization_settings
		SET default_alert_severity_threshold = COALESCE(NULLIF($2, ''), default_alert_severity_threshold),
		    auto_containment_enabled = $3,
		    auto_containment_threshold = COALESCE(NULLIF($4, 0), auto_containment_threshold),
		    allowed_ip_ranges = COALESCE($5, allowed_ip_ranges),
		    session_timeout_minutes = COALESCE(NULLIF($6, 0), session_timeout_minutes),
		    audit_logging_enabled = $7
		WHERE organization_id = $1
		RETURNING id
	`
)

// Valid organization roles per spec
var validRoles = []string{"owner", "admin", "analyst", "viewer", "automation_admin", "billing_admin"}

func NewCompanyRepository(db *db.DB) CompanyRepositoryInt {
	return &CompanyRepository{
		db: db,
	}
}

func (r *CompanyRepository) AddMemberToOrganization(ctx context.Context, organizationId string, userId string, roleId string) error {
	_, err := r.db.ExecContext(ctx, ADD_USER_TO_ORGANIZATION, organizationId, userId, roleId, "active")
	return err
}

func (r *CompanyRepository) GetIndustries(ctx context.Context) (*[]models.Industry, error) {
	var industries []models.Industry
	err := r.db.SelectContext(ctx, &industries, GET_INDUSTRIES)
	if err != nil {
		return nil, err
	}
	return &industries, nil
}

func (r *CompanyRepository) GetIndustryByID(ctx context.Context, id string) (*models.Industry, error) {
	var industry models.Industry
	err := r.db.GetContext(ctx, &industry, GET_INDUSTRY_BY_ID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("Industry not found")
		}
		return nil, err
	}
	return &industry, nil
}
func (r *CompanyRepository) InviteMemberToOrganization(ctx context.Context, req *requests.InviteMemberRequest, organizationId string, owner_id string, tokenHash *string, expiresAt time.Time) (*string, error) {
	var inviteId string
	err := r.db.QueryRowContext(
		ctx,
		INVITE_MEMBER_TO_ORGANIZATION,
		organizationId,
		req.Email,
		req.RoleId,
		owner_id,
		tokenHash,
		expiresAt,
	).Scan(&inviteId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invite already exists or user already invited")
		}
		return nil, err
	}

	return &inviteId, nil
}
func (r *CompanyRepository) GetByID(ctx context.Context, id string) (*models.OrganizationInvite, error) {
	var inv models.OrganizationInvite

	err := r.db.QueryRowContext(ctx, GET_INVITATION_BY_ID, id).Scan(
		&inv.ID,
		&inv.OrganizationID,
		&inv.Email,
		&inv.RoleID,
		&inv.Status,
		&inv.InvitedBy,
		&inv.TokenHash,
		&inv.ExpiresAt,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &inv, nil
}
func (r *CompanyRepository) GetMemberByEmail(ctx context.Context, email string, organizationId string) (*models.OrganizationMember, error) {
	var member models.OrganizationMember

	err := r.db.QueryRowContext(ctx, GET_MEMBER_BY_EMAIL, email, organizationId).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Email,
		&member.FirstName,
		&member.LastName,
		&member.AvatarURL,
		&member.Role,
		&member.Status,
		&member.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("Member not found with given email in this organization")
		}
		return nil, err
	}
	return &member, nil
}
func (r *CompanyRepository) MarkAccepted(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, MARK_INVITATION_ACCEPTED, id)
	return err
}
func (r *CompanyRepository) MarkExpired(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, MARK_INVITATION_EXPIRED, id)

	return err
}
func (r *CompanyRepository) GetOrganizationRoles(ctx context.Context) (*[]models.OrganizationRole, error) {
	var roles []models.OrganizationRole

	err := r.db.SelectContext(ctx, &roles, GET_ORGANIZATION_ROLES)
	if err != nil {
		return nil, err
	}

	return &roles, nil
}

func (r *CompanyRepository) GetOrganizationRoleByID(ctx context.Context, id string) (*models.OrganizationRole, error) {
	var role models.OrganizationRole

	err := r.db.GetContext(ctx, &role, GET_ORGANIZATION_ROLE_BY_ID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("Organization role not found")
		}
		return nil, err
	}

	return &role, nil
}

func (r *CompanyRepository) GetOrganizationByOwnerID(ctx context.Context, ownerId string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.GetContext(ctx, &org, GET_ORGANIZATION_BY_OWNER_ID, ownerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("Organization not found for owner")
		}
		return nil, err
	}
	return &org, nil
}

// GetOrganizationByID retrieves organization by ID
func (r *CompanyRepository) GetOrganizationByID(ctx context.Context, orgID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.GetContext(ctx, &org, getOrganizationByIDSQL, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("ORGANIZATION NOT FOUND")
		}
		return nil, err
	}
	return &org, nil
}

// UpdateOrganization updates organization metadata
func (r *CompanyRepository) UpdateOrganization(ctx context.Context, orgID string, req *requests.UpdateOrganizationRequest) error {
	var id string
	err := r.db.QueryRowContext(ctx, updateOrganizationSQL,
		orgID,
		req.Name,
		req.Industry,
		req.Country,
		req.Timezone,
		req.RiskThresholdDefault,
	).Scan(&id)
	return err
}

// ListMembers retrieves paginated members with filters
func (r *CompanyRepository) ListMembers(ctx context.Context, orgID string, page, pageSize int, role, search string) ([]models.OrganizationMember, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	var total int
	err := r.db.GetContext(ctx, &total, countMembersSQL, orgID, role, search)
	if err != nil {
		return nil, 0, err
	}

	var members []models.OrganizationMember
	err = r.db.SelectContext(ctx, &members, listMembersSQL, orgID, role, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetMemberByID retrieves a specific member
func (r *CompanyRepository) GetMemberByID(ctx context.Context, memberID, orgID string) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.GetContext(ctx, &member, getMemberByIDSQL, memberID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFoundError("MEMBER NOT FOUND")
		}
		return nil, err
	}
	return &member, nil
}

// AddMember adds a new member to the organization
func (r *CompanyRepository) AddMember(ctx context.Context, orgID, userID, role, invitedBy string) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.QueryRowContext(ctx, addMemberSQL, orgID, userID, role, invitedBy).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.Status,
		&member.InvitedBy,
		&member.InvitedAt,
		&member.CreatedAt,
		&member.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, apperrors.ConflictError("MEMBER ALREADY EXISTS")
		}
		return nil, err
	}
	return &member, nil
}

// UpdateMemberRole updates a member's role
func (r *CompanyRepository) UpdateMemberRole(ctx context.Context, memberID, orgID, newRole string) error {
	result, err := r.db.ExecContext(ctx, updateMemberRoleSQL, memberID, orgID, newRole)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("MEMBER NOT FOUND")
	}
	return nil
}

// UpdateMemberStatus updates the member's status in the organization.
func (r *CompanyRepository) UpdateMemberStatus(ctx context.Context, memberID, orgID, newStatus string) error {
	result, err := r.db.ExecContext(ctx, updateMemberStatusSQL, memberID, orgID, newStatus)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("MEMBER NOT FOUND")
	}
	return nil
}

// ResetMember2FA disables 2FA for a member in the organization.
func (r *CompanyRepository) ResetMember2FA(ctx context.Context, memberID, orgID string) error {
	result, err := r.db.ExecContext(ctx, resetMember2FASQL, memberID, orgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("MEMBER NOT FOUND")
	}
	return nil
}

// RemoveMember soft-deletes a member from the organization
func (r *CompanyRepository) RemoveMember(ctx context.Context, memberID, orgID string) error {
	result, err := r.db.ExecContext(ctx, removeMemberSQL, memberID, orgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("MEMBER NOT FOUND")
	}
	return nil
}

// GetMemberCountByRole counts members with a specific role
func (r *CompanyRepository) GetMemberCountByRole(ctx context.Context, orgID, role string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, countMembersByRoleSQL, orgID, role)
	return count, err
}

// GetOrganizationSettings retrieves org settings, creates defaults if not exists
func (r *CompanyRepository) GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error) {
	var settings models.OrganizationSettings
	err := r.db.GetContext(ctx, &settings, getOrganizationSettingsSQL, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.CreateOrganizationSettings(ctx, orgID)
		}
		return nil, err
	}
	return &settings, nil
}

// CreateOrganizationSettings creates default settings for an organization
func (r *CompanyRepository) CreateOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error) {
	var settings models.OrganizationSettings
	err := r.db.QueryRowContext(ctx, createOrganizationSettingsSQL, orgID).Scan(
		&settings.ID,
		&settings.OrganizationID,
		&settings.DefaultAlertSeverityThreshold,
		&settings.AutoContainmentEnabled,
		&settings.AutoContainmentThreshold,
		&settings.AllowedIPRanges,
		&settings.SessionTimeoutMinutes,
		&settings.AuditLoggingEnabled,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateOrganizationSettings updates org settings
func (r *CompanyRepository) UpdateOrganizationSettings(ctx context.Context, orgID string, req *requests.UpdateOrganizationSettingsRequest) error {
	ipRangesJSON := "[]"
	if len(req.AllowedIPRanges) > 0 {
		quoted := make([]string, len(req.AllowedIPRanges))
		for i, ip := range req.AllowedIPRanges {
			quoted[i] = fmt.Sprintf("%q", ip)
		}
		ipRangesJSON = "[" + strings.Join(quoted, ",") + "]"
	}

	var id string
	err := r.db.QueryRowContext(ctx, updateOrganizationSettingsSQL,
		orgID,
		req.DefaultAlertSeverityThreshold,
		req.AutoContainmentEnabled,
		req.AutoContainmentThreshold,
		ipRangesJSON,
		req.SessionTimeoutMinutes,
		req.AuditLoggingEnabled,
	).Scan(&id)
	return err
}

// GetValidRoles returns all valid organization roles
func (r *CompanyRepository) GetValidRoles() []string {
	return validRoles
}

// IsValidRole checks if a role is valid
func (r *CompanyRepository) IsValidRole(role string) bool {
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// Custom Roles & Permissions SQL
var (
	listPermissionsSQL = `
		SELECT id, name, description, category, resource, action, created_at, updated_at
		FROM permissions
		ORDER BY category, name
	`
	listPermissionGroupsSQL = `
		SELECT pg.id, pg.name, pg.description, pg.category, pg.created_at, pg.updated_at
		FROM permission_groups pg
		ORDER BY pg.category, pg.name
	`
	listCustomRolesSQL = `
		SELECT id, organization_id, name, description, is_system_role, created_at, updated_at
		FROM custom_roles
		WHERE organization_id = $1
		ORDER BY is_system_role DESC, name ASC
	`
	getCustomRoleByIDSQL = `
		SELECT id, organization_id, name, description, is_system_role, created_at, updated_at
		FROM custom_roles
		WHERE id = $1 AND organization_id = $2
	`
	createCustomRoleSQL = `
		INSERT INTO custom_roles (organization_id, name, description, is_system_role)
		VALUES ($1, $2, $3, false)
		RETURNING id, organization_id, name, description, is_system_role, created_at, updated_at
	`
	updateCustomRoleSQL = `
		UPDATE custom_roles
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $4
		RETURNING id, organization_id, name, description, is_system_role, created_at, updated_at
	`
	deleteCustomRoleSQL = `
		DELETE FROM custom_roles
		WHERE id = $1 AND organization_id = $2 AND is_system_role = false
	`
	getRolePermissionsSQL = `
		SELECT p.name
		FROM custom_role_permissions crp
		JOIN permissions p ON crp.permission_id = p.id
		WHERE crp.custom_role_id = $1
		ORDER BY p.name
	`
	getRolePermissionGroupsSQL = `
		SELECT pg.name
		FROM custom_role_permission_groups crpg
		JOIN permission_groups pg ON crpg.permission_group_id = pg.id
		WHERE crpg.custom_role_id = $1
		ORDER BY pg.name
	`
	setRolePermissionGroupsSQL = `
		DELETE FROM custom_role_permission_groups WHERE custom_role_id = $1
	`
	setRolePermissionsSQL = `
		DELETE FROM custom_role_permissions WHERE custom_role_id = $1
	`
)

// ListPermissions returns all permissions
func (r *CompanyRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	rows, err := r.db.QueryContext(ctx, listPermissionsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.Resource, &p.Action, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// ListPermissionGroups returns all permission groups
func (r *CompanyRepository) ListPermissionGroups(ctx context.Context) ([]models.PermissionGroup, error) {
	rows, err := r.db.QueryContext(ctx, listPermissionGroupsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.PermissionGroup
	for rows.Next() {
		var g models.PermissionGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Category, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// ListCustomRoles returns all custom roles for an organization
func (r *CompanyRepository) ListCustomRoles(ctx context.Context, orgID string) ([]models.CustomRole, error) {
	rows, err := r.db.QueryContext(ctx, listCustomRolesSQL, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.CustomRole
	for rows.Next() {
		var role models.CustomRole
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// GetCustomRoleByID returns a custom role by ID
func (r *CompanyRepository) GetCustomRoleByID(ctx context.Context, roleID, orgID string) (*models.CustomRole, error) {
	var role models.CustomRole
	err := r.db.QueryRowContext(ctx, getCustomRoleByIDSQL, roleID, orgID).Scan(
		&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("CUSTOM ROLE NOT FOUND")
		}
		return nil, err
	}
	return &role, nil
}

// CreateCustomRole creates a new custom role
func (r *CompanyRepository) CreateCustomRole(ctx context.Context, orgID string, req *requests.CreateCustomRoleRequest) (*models.CustomRole, error) {
	var role models.CustomRole
	err := r.db.QueryRowContext(ctx, createCustomRoleSQL, orgID, req.Name, req.Description).Scan(
		&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, apperrors.ConflictError("ROLE NAME ALREADY EXISTS")
		}
		return nil, err
	}

	// Set permission groups if provided
	if len(req.PermissionGroups) > 0 {
		if err := r.SetRolePermissionGroups(ctx, role.ID, req.PermissionGroups); err != nil {
			return nil, err
		}
	}

	// Set permissions if provided
	if len(req.Permissions) > 0 {
		if err := r.SetRolePermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return &role, nil
}

// UpdateCustomRole updates a custom role
func (r *CompanyRepository) UpdateCustomRole(ctx context.Context, roleID, orgID string, req *requests.UpdateCustomRoleRequest) error {
	var role models.CustomRole
	err := r.db.QueryRowContext(ctx, updateCustomRoleSQL, roleID, req.Name, req.Description, orgID).Scan(
		&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.NotFoundError("CUSTOM ROLE NOT FOUND")
		}
		return err
	}

	// Update permission groups if provided
	if req.PermissionGroups != nil {
		if err := r.SetRolePermissionGroups(ctx, roleID, req.PermissionGroups); err != nil {
			return err
		}
	}

	// Update permissions if provided
	if req.Permissions != nil {
		if err := r.SetRolePermissions(ctx, roleID, req.Permissions); err != nil {
			return err
		}
	}

	return nil
}

// DeleteCustomRole deletes a custom role
func (r *CompanyRepository) DeleteCustomRole(ctx context.Context, roleID, orgID string) error {
	result, err := r.db.ExecContext(ctx, deleteCustomRoleSQL, roleID, orgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("CUSTOM ROLE NOT FOUND OR CANNOT DELETE SYSTEM ROLE")
	}
	return nil
}

// GetRolePermissions returns permissions for a role
func (r *CompanyRepository) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, getRolePermissionsSQL, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}
	return permissions, nil
}

// GetRolePermissionGroups returns permission groups for a role
func (r *CompanyRepository) GetRolePermissionGroups(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, getRolePermissionGroupsSQL, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		groups = append(groups, name)
	}
	return groups, nil
}

// SetRolePermissionGroups sets permission groups for a role
func (r *CompanyRepository) SetRolePermissionGroups(ctx context.Context, roleID string, groupIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing
	if _, err := tx.ExecContext(ctx, setRolePermissionGroupsSQL, roleID); err != nil {
		return err
	}

	// Add new
	for _, groupID := range groupIDs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO custom_role_permission_groups (custom_role_id, permission_group_id)
			VALUES ($1, $2)
		`, roleID, groupID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SetRolePermissions sets permissions for a role
func (r *CompanyRepository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing
	if _, err := tx.ExecContext(ctx, setRolePermissionsSQL, roleID); err != nil {
		return err
	}

	// Add new
	for _, permID := range permissionIDs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO custom_role_permissions (custom_role_id, permission_id)
			VALUES ($1, $2)
		`, roleID, permID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *CompanyRepository) CheckMemberInOrganization(ctx context.Context, userID, orgID string) (bool, error) {
	var isMember bool
	err := r.db.GetContext(ctx, &isMember, checkMemberInOrganizationSQL, userID, orgID)
	if err != nil {
		return false, err
	}
	return isMember, nil
}

// CanManageMembers checks if role can invite/remove members
func (r *CompanyRepository) CanManageMembers(actorRole string) bool {
	return actorRole == "owner" || actorRole == "admin"
}

// CanUpdateSettings checks if role can update organization settings
func (r *CompanyRepository) CanUpdateSettings(actorRole string) bool {
	return actorRole == "owner" || actorRole == "admin"
}
