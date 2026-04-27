package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type CompanyServicesInt interface {
	// Industries & Roles
	GetIndustries() (*[]models.GetIndustriesResponse, error)
	GetOrganizationRoles() (*[]models.GetOrganizationRolesResponse, error)
	GetOrganizationRoleByID(id string) (*models.GetOrganizationRolesResponse, error)

	// Legacy Invitation Flow
	InviteUserToOrganization(ctx context.Context, req *requests.BulkInviteMembersRequest, ownerId string) error
	AcceptInvitation(ctx context.Context, inviteId string, token string, userId string) error

	// Organization Metadata
	GetOrganization(ctx context.Context, orgID string) (*models.GetOrganizationResponse, error)
	UpdateOrganization(ctx context.Context, orgID, actorRole string, req *requests.UpdateOrganizationRequest) error

	// Members
	ListMembers(ctx context.Context, orgID string, req *requests.ListMembersRequest) ([]*models.OrganizationMemberResponse, int, error)
	InviteMembers(ctx context.Context, orgID, invitedBy string, req *requests.InviteMembersRequest) (*InviteMembersResult, error)
	UpdateMemberRole(ctx context.Context, memberID, orgID, actorRole, newRole string) error
	RemoveMember(ctx context.Context, memberID, orgID, actorRole, actorUserID string) error

	// Settings
	GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettingsResponse, error)
	UpdateOrganizationSettings(ctx context.Context, orgID, actorRole string, req *requests.UpdateOrganizationSettingsRequest) error

	// Role validation helpers
	CanManageMembers(role string) bool
	CanUpdateSettings(role string) bool

	// Permissions & Custom Roles
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	ListPermissionGroups(ctx context.Context) ([]models.PermissionGroup, error)
	ListCustomRoles(ctx context.Context, orgID string) ([]models.CustomRoleResponse, error)
	GetCustomRole(ctx context.Context, roleID, orgID string) (*models.CustomRoleResponse, error)
	CreateCustomRole(ctx context.Context, orgID string, req *requests.CreateCustomRoleRequest) (*models.CustomRoleResponse, error)
	UpdateCustomRole(ctx context.Context, roleID, orgID string, req *requests.UpdateCustomRoleRequest) error
	DeleteCustomRole(ctx context.Context, roleID, orgID string) error
}

type InviteMembersResult struct {
	InvitedCount  int      `json:"invited_count"`
	SkippedCount  int      `json:"skipped_count"`
	InvalidEmails []string `json:"invalid_emails"`
}

type CompanyServices struct {
	companyRepo repositories.CompanyRepositoryInt
	userRepo    repositories.UsersRepositoryInt
	mailer      mailer.EmailClientInt
	redisClient *redis.Client
	config      *config.BaseConfig
}

func NewCompanyServices(
	companyRepo repositories.CompanyRepositoryInt,
	userRepo repositories.UsersRepositoryInt,
	mailer mailer.EmailClientInt,
	redisClient *redis.Client,
	config *config.BaseConfig,
) CompanyServicesInt {
	return &CompanyServices{
		companyRepo: companyRepo,
		userRepo:    userRepo,
		mailer:      mailer,
		redisClient: redisClient,
		config:      config,
	}
}

func (s *CompanyServices) GetIndustries() (*[]models.GetIndustriesResponse, error) {
	industries, err := s.companyRepo.GetIndustries(context.Background())
	if err != nil {
		return nil, err
	}
	response := make([]models.GetIndustriesResponse, len(*industries))
	for i, industry := range *industries {
		response[i] = industry.ToGetIndustriesResponse()
	}
	return &response, nil
}
func (s *CompanyServices) GetOrganizationRoles() (*[]models.GetOrganizationRolesResponse, error) {
	roles, err := s.companyRepo.GetOrganizationRoles(context.Background())
	if err != nil {
		return nil, err
	}
	response := make([]models.GetOrganizationRolesResponse, len(*roles))
	for i, role := range *roles {
		response[i] = role.ToGetOrganizationRolesResponse()
	}
	return &response, nil
}

func (s *CompanyServices) GetOrganizationRoleByID(id string) (*models.GetOrganizationRolesResponse, error) {
	role, err := s.companyRepo.GetOrganizationRoleByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	response := role.ToGetOrganizationRolesResponse()
	return &response, nil
}

func (s *CompanyServices) InviteUserToOrganization(ctx context.Context, req *requests.BulkInviteMembersRequest, ownerId string) error {
	organization, err := s.companyRepo.GetOrganizationByOwnerID(ctx, ownerId)
	if err != nil {
		return err
	}
	for _, invite := range req.Invites {
		member, err := s.companyRepo.GetMemberByEmail(ctx, invite.Email, organization.ID)
		if err != nil {
			var appErr *apperrors.ErrorResponse
			if !errors.As(err, &appErr) || appErr.Code != apperrors.ErrNotFound {
				return err
			}
		}

		if member != nil {
			continue
		}
		role, err := s.companyRepo.GetOrganizationRoleByID(ctx, invite.RoleId)
		if err != nil {
			return err
		}
		if role.Name == "owner" {
			return fmt.Errorf("cannot assign owner role to invitee")
		}

		token, err := utils.GenerateRandomStringForHashing(8)
		if err != nil {
			return err
		}

		hash, err := utils.HashRandomString(token)
		if err != nil {
			return err
		}

		expiresAt := time.Now().Add(48 * time.Hour)

		inviteId, err := s.companyRepo.InviteMemberToOrganization(ctx, &invite, organization.ID, ownerId, &hash, expiresAt)
		if err != nil {
			return err
		}

		inviteLink := fmt.Sprintf("%s/company/invitations/accept?invite_id=%s&token=%s", s.config.FrontendBaseURL, *inviteId, token)

		redisKey := fmt.Sprintf("invite:%s", *inviteId)
		if err := s.redisClient.Set(ctx, redisKey, token, 48*time.Hour).Err(); err != nil {
			return err
		}

		if err := s.mailer.SendMemberInvitationEmail([]string{invite.Email}, mailer.MemberInvitationEmailData{
			OrganizationName: organization.Name,
			Role:             role.Name,
			InviteLink:       inviteLink,
		}); err != nil {
			return err
		}
	}

	return nil
}
func (s *CompanyServices) AcceptInvitation(ctx context.Context, inviteId string, token string, userId string) error {
	redisKey := fmt.Sprintf("invite:%s", inviteId)
	storedToken, err := s.redisClient.Get(ctx, redisKey).Result()

	if err == redis.Nil {
		return apperrors.BadException("invitation has expired or is invalid")
	}
	if err != nil {
		return err
	}
	if storedToken != token {
		return apperrors.UnauthorizedException("invalid invitation token")
	}

	invite, err := s.companyRepo.GetByID(ctx, inviteId)
	if err != nil {
		return err
	}
	if time.Now().After(invite.ExpiresAt) {
		return apperrors.BadException("invitation has expired")
	}

	if err := s.companyRepo.AddMemberToOrganization(ctx, invite.OrganizationID, userId, *invite.RoleID); err != nil {
		return err
	}

	if err := s.redisClient.Del(ctx, redisKey).Err(); err != nil {
		return err
	}

	return nil
}

// GetOrganization returns organization details
func (s *CompanyServices) GetOrganization(ctx context.Context, orgID string) (*models.GetOrganizationResponse, error) {
	org, err := s.companyRepo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return org.ToResponse(), nil
}

// UpdateOrganization updates organization metadata (owner/admin only)
func (s *CompanyServices) UpdateOrganization(ctx context.Context, orgID, actorRole string, req *requests.UpdateOrganizationRequest) error {
	if !s.CanUpdateSettings(actorRole) {
		return apperrors.NewCustomError("ONLY OWNER OR ADMIN CAN UPDATE ORGANIZATION", fiber.StatusForbidden, "FORBIDDEN")
	}
	if req.RiskThresholdDefault < 0 || req.RiskThresholdDefault > 100 {
		return apperrors.BadException("RISK_THRESHOLD_MUST_BE_BETWEEN_0_AND_100")
	}
	return s.companyRepo.UpdateOrganization(ctx, orgID, req)
}

// ListMembers returns paginated member list
func (s *CompanyServices) ListMembers(ctx context.Context, orgID string, req *requests.ListMembersRequest) ([]*models.OrganizationMemberResponse, int, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 25
	}
	role := strings.ToLower(req.Role)
	if role != "" && !s.companyRepo.IsValidRole(role) {
		role = ""
	}

	members, total, err := s.companyRepo.ListMembers(ctx, orgID, req.Page, req.PageSize, role, req.Search)
	if err != nil {
		return nil, 0, err
	}

	var responses []*models.OrganizationMemberResponse
	for _, m := range members {
		responses = append(responses, m.ToResponse())
	}
	return responses, total, nil
}

// InviteMembers invites users to join the organization
func (s *CompanyServices) InviteMembers(ctx context.Context, orgID, invitedBy string, req *requests.InviteMembersRequest) (*InviteMembersResult, error) {
	result := &InviteMembersResult{
		InvitedCount:  0,
		SkippedCount:  0,
		InvalidEmails: []string{},
	}

	for _, invite := range req.Invites {
		email := strings.ToLower(strings.TrimSpace(invite.Email))
		if _, err := mail.ParseAddress(email); err != nil {
			result.InvalidEmails = append(result.InvalidEmails, email)
			continue
		}

		user, err := s.userRepo.GetUserByEmail(ctx, email)
		if err != nil || user == nil {
			result.SkippedCount++
			continue
		}

		_, err = s.companyRepo.AddMember(ctx, orgID, user.ID, invite.Role, invitedBy)
		if err != nil {
			if err.Error() == "MEMBER ALREADY EXISTS" || strings.Contains(err.Error(), "duplicate") {
				result.SkippedCount++
				continue
			}
			result.SkippedCount++
			continue
		}
		result.InvitedCount++
		// TODO: Send invitation email with invite.FullName, invite.Department, invite.ForceMFA, invite.Message
	}

	return result, nil
}

// UpdateMemberRole updates a member's role
func (s *CompanyServices) UpdateMemberRole(ctx context.Context, memberID, orgID, actorRole, newRole string) error {
	if !s.CanManageMembers(actorRole) {
		return apperrors.NewCustomError("ONLY OWNER OR ADMIN CAN UPDATE MEMBER ROLES", fiber.StatusForbidden, "FORBIDDEN")
	}
	if !s.companyRepo.IsValidRole(newRole) {
		return apperrors.BadException("INVALID_ROLE")
	}
	if newRole == "owner" {
		return apperrors.NewCustomError("CANNOT_ASSIGN_OWNER_ROLE", fiber.StatusForbidden, "FORBIDDEN")
	}

	member, err := s.companyRepo.GetMemberByID(ctx, memberID, orgID)
	if err != nil {
		return err
	}
	if member.Role == "owner" && actorRole != "owner" {
		return apperrors.NewCustomError("ONLY_OWNER_CAN_MODIFY_OWNER", fiber.StatusForbidden, "FORBIDDEN")
	}

	return s.companyRepo.UpdateMemberRole(ctx, memberID, orgID, newRole)
}

// RemoveMember removes a member from the organization
func (s *CompanyServices) RemoveMember(ctx context.Context, memberID, orgID, actorRole, actorUserID string) error {
	if !s.CanManageMembers(actorRole) {
		return apperrors.NewCustomError("ONLY OWNER OR ADMIN CAN REMOVE MEMBERS", fiber.StatusForbidden, "FORBIDDEN")
	}

	member, err := s.companyRepo.GetMemberByID(ctx, memberID, orgID)
	if err != nil {
		return err
	}

	if member.UserID == actorUserID {
		ownerCount, err := s.companyRepo.GetMemberCountByRole(ctx, orgID, "owner")
		if err != nil {
			return err
		}
		if ownerCount <= 1 && member.Role == "owner" {
			return apperrors.NewCustomError("CANNOT_REMOVE_LAST_OWNER", fiber.StatusForbidden, "FORBIDDEN")
		}
	}

	if member.Role == "owner" && actorRole != "owner" {
		return apperrors.NewCustomError("ONLY_OWNER_CAN_REMOVE_OWNER", fiber.StatusForbidden, "FORBIDDEN")
	}

	return s.companyRepo.RemoveMember(ctx, memberID, orgID)
}

// GetOrganizationSettings returns org settings
func (s *CompanyServices) GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettingsResponse, error) {
	settings, err := s.companyRepo.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return settings.ToResponse(), nil
}

// UpdateOrganizationSettings updates org settings
func (s *CompanyServices) UpdateOrganizationSettings(ctx context.Context, orgID, actorRole string, req *requests.UpdateOrganizationSettingsRequest) error {
	if !s.CanUpdateSettings(actorRole) {
		return apperrors.NewCustomError("ONLY OWNER OR ADMIN CAN UPDATE SETTINGS", fiber.StatusForbidden, "FORBIDDEN")
	}

	validSeverities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if req.DefaultAlertSeverityThreshold != "" && !validSeverities[req.DefaultAlertSeverityThreshold] {
		return apperrors.BadException("INVALID_SEVERITY_THRESHOLD")
	}
	if req.AutoContainmentThreshold < 0 || req.AutoContainmentThreshold > 100 {
		return apperrors.BadException("CONTAINMENT_THRESHOLD_MUST_BE_BETWEEN_0_AND_100")
	}
	if req.SessionTimeoutMinutes < 5 || req.SessionTimeoutMinutes > 1440 {
		return apperrors.BadException("SESSION_TIMEOUT_MUST_BE_BETWEEN_5_AND_1440_MINUTES")
	}

	return s.companyRepo.UpdateOrganizationSettings(ctx, orgID, req)
}

// CanManageMembers checks if role can manage members
func (s *CompanyServices) CanManageMembers(role string) bool {
	return s.companyRepo.CanManageMembers(role)
}

// CanUpdateSettings checks if role can update settings
func (s *CompanyServices) CanUpdateSettings(role string) bool {
	return s.companyRepo.CanUpdateSettings(role)
}

// ListPermissions returns all permissions
func (s *CompanyServices) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.companyRepo.ListPermissions(ctx)
}

// ListPermissionGroups returns all permission groups
func (s *CompanyServices) ListPermissionGroups(ctx context.Context) ([]models.PermissionGroup, error) {
	return s.companyRepo.ListPermissionGroups(ctx)
}

// ListCustomRoles returns all custom roles for an organization
func (s *CompanyServices) ListCustomRoles(ctx context.Context, orgID string) ([]models.CustomRoleResponse, error) {
	roles, err := s.companyRepo.ListCustomRoles(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var responses []models.CustomRoleResponse
	for _, role := range roles {
		resp := models.CustomRoleResponse{
			ID:           role.ID,
			Name:         role.Name,
			Description:  role.Description,
			IsSystemRole: role.IsSystemRole,
			CreatedAt:    role.CreatedAt,
			UpdatedAt:    role.UpdatedAt,
		}

		// Get permission groups
		groups, _ := s.companyRepo.GetRolePermissionGroups(ctx, role.ID)
		resp.PermissionGroups = groups

		// Get permissions
		permissions, _ := s.companyRepo.GetRolePermissions(ctx, role.ID)
		resp.Permissions = permissions

		responses = append(responses, resp)
	}

	return responses, nil
}

// GetCustomRole returns a custom role by ID
func (s *CompanyServices) GetCustomRole(ctx context.Context, roleID, orgID string) (*models.CustomRoleResponse, error) {
	role, err := s.companyRepo.GetCustomRoleByID(ctx, roleID, orgID)
	if err != nil {
		return nil, err
	}

	resp := &models.CustomRoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		IsSystemRole: role.IsSystemRole,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
	}

	// Get permission groups
	groups, _ := s.companyRepo.GetRolePermissionGroups(ctx, role.ID)
	resp.PermissionGroups = groups

	// Get permissions
	permissions, _ := s.companyRepo.GetRolePermissions(ctx, role.ID)
	resp.Permissions = permissions

	return resp, nil
}

// CreateCustomRole creates a new custom role
func (s *CompanyServices) CreateCustomRole(ctx context.Context, orgID string, req *requests.CreateCustomRoleRequest) (*models.CustomRoleResponse, error) {
	role, err := s.companyRepo.CreateCustomRole(ctx, orgID, req)
	if err != nil {
		return nil, err
	}

	resp := &models.CustomRoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		IsSystemRole: role.IsSystemRole,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
	}

	// Get permission groups
	groups, _ := s.companyRepo.GetRolePermissionGroups(ctx, role.ID)
	resp.PermissionGroups = groups

	// Get permissions
	permissions, _ := s.companyRepo.GetRolePermissions(ctx, role.ID)
	resp.Permissions = permissions

	return resp, nil
}

// UpdateCustomRole updates a custom role
func (s *CompanyServices) UpdateCustomRole(ctx context.Context, roleID, orgID string, req *requests.UpdateCustomRoleRequest) error {
	return s.companyRepo.UpdateCustomRole(ctx, roleID, orgID, req)
}

// DeleteCustomRole deletes a custom role
func (s *CompanyServices) DeleteCustomRole(ctx context.Context, roleID, orgID string) error {
	return s.companyRepo.DeleteCustomRole(ctx, roleID, orgID)
}
