package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/domain"
	"sage-backend/internal/users/ports/inbound"
	"sage-backend/internal/users/ports/outbound"
	"sage-backend/internal/users/usecase/dto"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)



type CompanyServices struct {
	companyRepo outbound.CompanyRepository
	userRepo    outbound.UserRepository
	mailer      mailer.EmailClientInt
	redisClient *redis.Client
	config      *config.BaseConfig
}

func NewCompanyServices(
	companyRepo outbound.CompanyRepository,
	userRepo outbound.UserRepository,
	mailer mailer.EmailClientInt,
	redisClient *redis.Client,
	config *config.BaseConfig,
) inbound.CompanyUseCase {
	return &CompanyServices{
		companyRepo: companyRepo,
		userRepo:    userRepo,
		mailer:      mailer,
		redisClient: redisClient,
		config:      config,
	}
}

func (s *CompanyServices) GetIndustries(ctx context.Context) (*[]domain.GetIndustriesResponse, error) {
	industries, err := s.companyRepo.GetIndustries(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]domain.GetIndustriesResponse, len(*industries))
	for i, industry := range *industries {
		response[i] = industry.ToGetIndustriesResponse()
	}
	return &response, nil
}
func (s *CompanyServices) GetOrganizationRoles(ctx context.Context) (*[]domain.GetOrganizationRolesResponse, error) {
	roles, err := s.companyRepo.GetOrganizationRoles(ctx)
	if err != nil {
		return nil, err
	}
	response := make([]domain.GetOrganizationRolesResponse, len(*roles))
	for i, role := range *roles {
		response[i] = role.ToGetOrganizationRolesResponse()
	}
	return &response, nil
}

func (s *CompanyServices) GetOrganizationRoleByID(ctx context.Context, id string) (*domain.GetOrganizationRolesResponse, error) {
	role, err := s.companyRepo.GetOrganizationRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := role.ToGetOrganizationRolesResponse()
	return &response, nil
}

func (s *CompanyServices) InviteUserToOrganization(ctx context.Context, req *dto.BulkInviteMembersRequest, ownerId string) error {
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
func (s *CompanyServices) GetOrganization(ctx context.Context, orgID string) (*domain.GetOrganizationResponse, error) {
	org, err := s.companyRepo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return org.ToResponse(), nil
}

// UpdateOrganization updates organization metadata (owner/admin only)
func (s *CompanyServices) UpdateOrganization(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationRequest) error {
	if !s.CanUpdateSettings(actorRole) {
		return apperrors.NewCustomError("ONLY OWNER OR ADMIN CAN UPDATE ORGANIZATION", fiber.StatusForbidden, "FORBIDDEN")
	}
	if req.RiskThresholdDefault < 0 || req.RiskThresholdDefault > 100 {
		return apperrors.BadException("RISK_THRESHOLD_MUST_BE_BETWEEN_0_AND_100")
	}
	return s.companyRepo.UpdateOrganization(ctx, orgID, req)
}

// ListMembers returns paginated member list
func (s *CompanyServices) ListMembers(ctx context.Context, orgID string, req *dto.ListMembersRequest) ([]*domain.OrganizationMemberResponse, int, error) {
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

	var responses []*domain.OrganizationMemberResponse
	for _, m := range members {
		responses = append(responses, m.ToResponse())
	}
	return responses, total, nil
}

// InviteMembers invites users to join the organization
func (s *CompanyServices) InviteMembers(ctx context.Context, orgID, invitedBy string, req *dto.InviteMembersRequest) (*dto.InviteMembersResult, error) {
	result := &dto.InviteMembersResult{
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
func (s *CompanyServices) GetOrganizationSettings(ctx context.Context, orgID string) (*domain.OrganizationSettingsResponse, error) {
	settings, err := s.companyRepo.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return settings.ToResponse(), nil
}

// UpdateOrganizationSettings updates org settings
func (s *CompanyServices) UpdateOrganizationSettings(ctx context.Context, orgID, actorRole string, req *dto.UpdateOrganizationSettingsRequest) error {
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
func (s *CompanyServices) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.companyRepo.ListPermissions(ctx)
}

// ListPermissionGroups returns all permission groups
func (s *CompanyServices) ListPermissionGroups(ctx context.Context) ([]domain.PermissionGroup, error) {
	return s.companyRepo.ListPermissionGroups(ctx)
}

// ListCustomRoles returns all custom roles for an organization
func (s *CompanyServices) ListCustomRoles(ctx context.Context, orgID string) ([]domain.CustomRoleResponse, error) {
	roles, err := s.companyRepo.ListCustomRoles(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var responses []domain.CustomRoleResponse
	for _, role := range roles {
		resp := domain.CustomRoleResponse{
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
func (s *CompanyServices) GetCustomRole(ctx context.Context, roleID, orgID string) (*domain.CustomRoleResponse, error) {
	role, err := s.companyRepo.GetCustomRoleByID(ctx, roleID, orgID)
	if err != nil {
		return nil, err
	}

	resp := &domain.CustomRoleResponse{
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
func (s *CompanyServices) CreateCustomRole(ctx context.Context, orgID string, req *dto.CreateCustomRoleRequest) (*domain.CustomRoleResponse, error) {
	role, err := s.companyRepo.CreateCustomRole(ctx, orgID, req)
	if err != nil {
		return nil, err
	}

	resp := &domain.CustomRoleResponse{
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
func (s *CompanyServices) UpdateCustomRole(ctx context.Context, roleID, orgID string, req *dto.UpdateCustomRoleRequest) error {
	return s.companyRepo.UpdateCustomRole(ctx, roleID, orgID, req)
}

// DeleteCustomRole deletes a custom role
func (s *CompanyServices) DeleteCustomRole(ctx context.Context, roleID, orgID string) error {
	return s.companyRepo.DeleteCustomRole(ctx, roleID, orgID)
}
