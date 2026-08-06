package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/identity/domain"
	"sage-backend/internal/identity/ports/inbound"
	"sage-backend/internal/identity/ports/outbound"
	"sage-backend/internal/identity/usecase/dto"
	orgDomain "sage-backend/internal/organization/domain"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserServices struct {
	userRepo    outbound.UserRepository
	profileRepo outbound.ProfileRepository
	redisClient *redis.Client
	uploader    *s3.Uploader
}

func NewUsersServices(usersRepo outbound.UserRepository, profileRepo outbound.ProfileRepository, redisClient *redis.Client, uploader *s3.Uploader) inbound.UserUseCase {
	return &UserServices{
		userRepo:    usersRepo,
		profileRepo: profileRepo,
		redisClient: redisClient,
		uploader:    uploader,
	}
}

// GetProfile returns basic user profile (legacy method)
func (s *UserServices) GetProfile(ctx context.Context, userID string) (*dto.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	orgs, err := s.userRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := dto.UserToResponse(user, orgs)

	if resp.AvatarURL != "" {
		presignedUrl, err := s.uploader.GenerateSignedURL(ctx, resp.AvatarURL)
		if err != nil {
			return nil, err
		}
		resp.AvatarURL = presignedUrl
	}
	return resp, nil
}

// UpdateProfile updates basic profile fields (legacy method)
func (s *UserServices) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) error {
	return s.userRepo.UpdateUser(ctx, userID, req)
}

// GetIdentity returns the comprehensive identity profile for /api/v1/profile
func (s *UserServices) GetIdentity(ctx context.Context, userID, orgID string) (*dto.ProfileResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var org *orgDomain.Organization
	if orgID != "" {
		orgs, err := s.userRepo.GetUserOrganizations(ctx, userID)
		if err == nil && orgs != nil {
			for _, o := range *orgs {
				if o.ID == orgID {
					org = &o
					break
				}
			}
		}
	}

	resp := dto.UserToProfileResponse(user, org)
	if resp.AvatarURL != "" {
		presignedUrl, err := s.uploader.GenerateSignedURL(ctx, resp.AvatarURL)
		if err != nil {
			return nil, err
		}
		resp.AvatarURL = presignedUrl
	}
	return resp, nil
}

// UpdateIdentity updates user identity fields (full_name, phone_number, backup_email, avatar_url)
func (s *UserServices) UpdateIdentity(ctx context.Context, userID string, req *dto.UpdateIdentityRequest) error {
	if req.FullName != "" {
		parts := strings.SplitN(req.FullName, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}
		err := s.userRepo.UpdateUser(ctx, userID, &dto.UpdateProfileRequest{
			FirstName: firstName,
			LastName:  lastName,
		})
		if err != nil {
			return err
		}
	}

	if req.PhoneNumber != "" || req.BackupEmail != "" {
		if err := s.userRepo.UpdateUserContactInfo(ctx, userID, req.PhoneNumber, req.BackupEmail); err != nil {
			return err
		}
	}

	if req.AvatarURL != nil {
		if err := s.profileRepo.UpdateUserAvatar(ctx, userID, *req.AvatarURL); err != nil {
			return err
		}
	}

	return nil
}

// GetPreferences returns user preferences for /api/v1/profile/preferences
func (s *UserServices) GetPreferences(ctx context.Context, userID, orgID string) (*domain.UserPreferences, error) {
	return s.profileRepo.GetUserPreferences(ctx, userID, orgID)
}

// UpdatePreferences updates user preferences
func (s *UserServices) UpdatePreferences(ctx context.Context, userID, orgID string, req *dto.UpdatePreferencesRequest) error {
	prefs := &domain.UserPreferences{
		UserID:               userID,
		Theme:                req.Theme,
		Timezone:             req.Timezone,
		Language:             req.Language,
		DateFormat:           req.DateFormat,
		DashboardDefaultView: req.DashboardDefaultView,
		TablePageSize:        req.TablePageSize,
		AutoRefreshInterval:  req.AutoRefreshInterval,
	}
	return s.profileRepo.UpsertUserPreferences(ctx, userID, orgID, prefs)
}

// GetNotifications returns user notification settings for /api/v1/profile/notifications
func (s *UserServices) GetNotifications(ctx context.Context, userID, orgID string) (*domain.UserNotifications, error) {
	return s.profileRepo.GetUserNotifications(ctx, userID, orgID)
}

// UpdateNotifications updates user notification settings
func (s *UserServices) UpdateNotifications(ctx context.Context, userID, orgID string, req *dto.UpdateNotificationsRequest) error {
	notifs := &domain.UserNotifications{
		UserID:                    userID,
		EmailEnabled:              req.EmailEnabled,
		PushEnabled:               req.PushEnabled,
		SlackEnabled:              req.SlackEnabled,
		ProductUpdates:            req.ProductUpdates,
		WeeklySummary:             req.WeeklySummary,
		AlertSeverityThreshold:    req.AlertSeverityThreshold,
		NotifyOnNewAlert:          req.NotifyOnNewAlert,
		NotifyOnIncidentUpdate:    req.NotifyOnIncidentUpdate,
		NotifyOnPlaybookExecution: req.NotifyOnPlaybookExecution,
	}
	return s.profileRepo.UpsertUserNotifications(ctx, userID, orgID, notifs)
}

// GetSessions returns all active sessions for /api/v1/profile/sessions
func (s *UserServices) GetSessions(ctx context.Context, userID, orgID, currentSessionID string) ([]*dto.UserSessionResponse, error) {
	sessions, err := s.profileRepo.GetUserSessions(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	var responses []*dto.UserSessionResponse
	for _, sess := range sessions {
		responses = append(responses, dto.UserSessionToResponse(&sess, currentSessionID))
	}
	return responses, nil
}

// RevokeSession revokes a specific session for /api/v1/profile/sessions/{id}
func (s *UserServices) RevokeSession(ctx context.Context, sessionID, userID string) error {
	return s.profileRepo.RevokeUserSession(ctx, sessionID, userID)
}

// CreateSession creates a new user session
func (s *UserServices) CreateSession(ctx context.Context, userID, orgID, tokenHash, ipAddress, userAgent string) (*domain.UserSession, error) {
	now := time.Now()
	session := domain.NewUserSession(
		"",
		userID,
		sql.NullString{String: orgID, Valid: orgID != ""},
		tokenHash,
		sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		sql.NullString{String: userAgent, Valid: userAgent != ""},
		sql.NullString{},
		sql.NullString{},
		false,
		now.Add(24*time.Hour),
		now,
	)

	if err := s.profileRepo.CreateUserSession(ctx, session); err != nil {
		return nil, err
	}

	// Update last login time
	_ = s.profileRepo.UpdateLastLogin(ctx, userID)

	return session, nil
}

// GetActivity returns paginated user activity for /api/v1/profile/activity
func (s *UserServices) GetActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]domain.AuditLog, int, error) {
	return s.profileRepo.GetUserActivityLog(ctx, userID, orgID, page, pageSize)
}

// UploadAvatar uploads the provided file to S3 and updates the user's avatar URL
func (s *UserServices) UploadAvatar(ctx context.Context, userID string, file multipart.File, mimeType string) (string, string, error) {
	if s.uploader == nil {
		return "", "", errors.New("uploader not configured")
	}

	url, key, err := s.uploader.UploadAvatar(ctx, file, userID, mimeType)
	if err != nil {
		return "", "", err
	}

	if err := s.profileRepo.UpdateUserAvatar(ctx, userID, key); err != nil {
		return "", "", err
	}

	return url, key, nil
}

// LogActivity creates an audit log entry
func (s *UserServices) LogActivity(ctx context.Context, userID, orgID, actionType, resourceType, resourceID string, metadata map[string]interface{}, ipAddress, userAgent string) error {
	log := &domain.AuditLog{
		UserID:     userID,
		ActionType: actionType,
	}

	if orgID != "" {
		log.OrganizationID = sql.NullString{String: orgID, Valid: true}
	}
	if resourceType != "" {
		log.ResourceType = sql.NullString{String: resourceType, Valid: true}
	}
	if resourceID != "" {
		log.ResourceID = sql.NullString{String: resourceID, Valid: true}
	}
	if ipAddress != "" {
		log.IPAddress = sql.NullString{String: ipAddress, Valid: true}
	}
	if userAgent != "" {
		log.UserAgent = sql.NullString{String: userAgent, Valid: true}
	}
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			log.Metadata = sql.NullString{String: string(b), Valid: true}
		}
	}

	return s.profileRepo.CreateAuditLog(ctx, log)
}
