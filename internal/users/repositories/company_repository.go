package repositories

import (
	"context"
	"database/sql"
	"errors"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/requests"
	"time"
)

type CompanyRepositoryInt interface {
	GetIndustries(ctx context.Context) (*[]models.Industry, error)
	GetOrganizationRoles(ctx context.Context) (*[]models.OrganizationRole, error)
	GetOrganizationRoleByID(ctx context.Context, id string) (*models.OrganizationRole, error)
	GetIndustryByID(ctx context.Context, id string) (*models.Industry, error)
	InviteMemberToOrganization(ctx context.Context, req *requests.InviteMemberRequest, organizationId string, owner_id string, tokenHash *string, expiresAt time.Time) (*string, error)
	GetByID(ctx context.Context, id string) (*models.OrganizationInvite, error)
	MarkAccepted(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	GetMemberByEmail(ctx context.Context, email string, organizationId string) (*models.OrganizationMember, error)
	GetOrganizationByOwnerID(ctx context.Context, ownerId string) (*models.Organization, error)
	AddMemberToOrganization(ctx context.Context, organizationId string, userId string, roleId string) error
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
		    u.first_name || ' ' || u.last_name AS name,
		    r.name AS role,
		    om.status,
		    om.created_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		LEFT JOIN organization_roles r ON r.id = om.role_id
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
)

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
		&member.Name,
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
