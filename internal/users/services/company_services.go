package services

import (
	"context"
	"errors"
	"fmt"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/mailer"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"time"

	"github.com/redis/go-redis/v9"
)

type CompanyServicesInt interface {
	GetIndustries() (*[]models.GetIndustriesResponse, error)
	GetOrganizationRoles() (*[]models.GetOrganizationRolesResponse, error)
	GetOrganizationRoleByID(id string) (*models.GetOrganizationRolesResponse, error)
	InviteUserToOrganization(ctx context.Context, req *requests.BulkInviteMembersRequest, ownerId string) error
	AcceptInvitation(ctx context.Context, inviteId string, token string, userId string) error
}

type CompanyServices struct {
	companyRepo repositories.CompanyRepositoryInt
	mailer      mailer.EmailClientInt
	redisClient *redis.Client
	config      *config.BaseConfig
}

func NewCompanyServices(
	companyRepo repositories.CompanyRepositoryInt,
	mailer mailer.EmailClientInt,
	redisClient *redis.Client,
	config *config.BaseConfig,
) CompanyServicesInt {
	return &CompanyServices{
		companyRepo: companyRepo,
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
