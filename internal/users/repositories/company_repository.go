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
	RemoveMember(ctx context.Context, memberID, orgID string) error
	GetMemberCountByRole(ctx context.Context, orgID, role string) (int, error)

	// Settings
	GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error)
	CreateOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettings, error)
	UpdateOrganizationSettings(ctx context.Context, orgID string, req *requests.UpdateOrganizationSettingsRequest) error

	// Role helpers
	GetValidRoles() []string
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
		    om.role,
		    om.status,
		    om.created_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
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
		    om.role::text,
		    om.department,
		    om.status::text,
		    om.invited_by,
		    om.invited_at,
		    om.joined_at,
		    u.last_login_at,
		    om.created_at,
		    om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		  AND ($2 = '' OR om.role::text = $2)
		  AND ($3 = '' OR u.email ILIKE '%' || $3 || '%' OR u.first_name ILIKE '%' || $3 || '%' OR u.last_name ILIKE '%' || $3 || '%')
		  AND om.status != 'removed'
		ORDER BY om.created_at DESC
		LIMIT $4 OFFSET $5
	`
	countMembersSQL = `
		SELECT COUNT(*)
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		  AND ($2 = '' OR om.role::text = $2)
		  AND ($3 = '' OR u.email ILIKE '%' || $3 || '%' OR u.first_name ILIKE '%' || $3 || '%' OR u.last_name ILIKE '%' || $3 || '%')
		  AND om.status != 'removed'
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
		    om.role::text,
		    om.department,
		    om.status::text,
		    om.invited_by,
		    om.invited_at,
		    om.joined_at,
		    u.last_login_at,
		    om.created_at,
		    om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.id = $1 AND om.organization_id = $2
	`
	addMemberSQL = `
		INSERT INTO organization_members (organization_id, user_id, role, status, invited_by, invited_at)
		VALUES ($1, $2, $3::organization_role, 'invited', $4, NOW())
		RETURNING id, organization_id, user_id, role::text, status::text, invited_by, invited_at, created_at, updated_at
	`
	updateMemberRoleSQL = `
		UPDATE organization_members
		SET role = $3::organization_role,
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id
	`
	removeMemberSQL = `
		UPDATE organization_members
		SET status = 'removed',
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id
	`
	countMembersByRoleSQL = `
		SELECT COUNT(*)
		FROM organization_members
		WHERE organization_id = $1 AND role::text = $2 AND status != 'removed'
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
var validRoles = []string{"owner", "admin", "analyst", "viewer"}

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

// CanManageMembers checks if role can invite/remove members
func (r *CompanyRepository) CanManageMembers(actorRole string) bool {
	return actorRole == "owner" || actorRole == "admin"
}

// CanUpdateSettings checks if role can update organization settings
func (r *CompanyRepository) CanUpdateSettings(actorRole string) bool {
	return actorRole == "owner" || actorRole == "admin"
}
