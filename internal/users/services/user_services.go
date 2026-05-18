package services

import (
	"context"
	"errors"
	"mime/multipart"
	"sage-backend/internal/shared/storage/s3"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/repositories"
	"sage-backend/internal/users/requests"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserServicesInt interface {
	// Basic profile (legacy)
	GetProfile(ctx context.Context, userID string) (*models.GetUserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *requests.UpdateProfileRequest) error

	// Identity & Access - /api/v1/profile
	GetIdentity(ctx context.Context, userID, orgID string) (*models.ProfileResponse, error)
	UpdateIdentity(ctx context.Context, userID string, req *requests.UpdateIdentityRequest) error

	// Preferences - /api/v1/profile/preferences
	GetPreferences(ctx context.Context, userID, orgID string) (*models.UserPreferencesResponse, error)
	UpdatePreferences(ctx context.Context, userID, orgID string, req *requests.UpdatePreferencesRequest) error

	// Notifications - /api/v1/profile/notifications
	GetNotifications(ctx context.Context, userID, orgID string) (*models.UserNotificationsResponse, error)
	UpdateNotifications(ctx context.Context, userID, orgID string, req *requests.UpdateNotificationsRequest) error

	// Sessions - /api/v1/profile/sessions
	GetSessions(ctx context.Context, userID, orgID, currentSessionID string) ([]*models.UserSessionResponse, error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	CreateSession(ctx context.Context, userID, orgID, tokenHash, ipAddress, userAgent string) (*models.UserSession, error)

	// Activity - /api/v1/profile/activity
	GetActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]*models.UserActivityResponse, int, error)
	LogActivity(ctx context.Context, userID, orgID, actionType, resourceType, resourceID string, metadata map[string]interface{}, ipAddress, userAgent string) error
	// Storage
	UploadAvatar(ctx context.Context, userID string, file multipart.File, mimeType string) (string, string, error)
}

type UserServices struct {
	userRepo    repositories.UsersRepositoryInt
	profileRepo repositories.ProfileRepositoryInt
	redisClient *redis.Client
	uploader    *s3.Uploader
}

func NewUsersServices(usersRepo repositories.UsersRepositoryInt, profileRepo repositories.ProfileRepositoryInt, redisClient *redis.Client, uploader *s3.Uploader) UserServicesInt {
	return &UserServices{
		userRepo:    usersRepo,
		profileRepo: profileRepo,
		redisClient: redisClient,
		uploader:    uploader,
	}
}

// GetProfile returns basic user profile (legacy method)
func (s *UserServices) GetProfile(ctx context.Context, userID string) (*models.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	orgs, err := s.userRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := user.ToResponse(orgs)

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
func (s *UserServices) UpdateProfile(ctx context.Context, userID string, req *requests.UpdateProfileRequest) error {
	return s.userRepo.UpdateUser(ctx, userID, req)
}

// GetIdentity returns the comprehensive identity profile for /api/v1/profile
func (s *UserServices) GetIdentity(ctx context.Context, userID, orgID string) (*models.ProfileResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var org *models.Organization
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

	resp := user.ToProfileResponse(org)
	if resp.AvatarURL != "" {
		presignedUrl, err := s.uploader.GenerateSignedURL(ctx, resp.AvatarURL)
		if err != nil {
			return nil, err
		}
		resp.AvatarURL = presignedUrl
	}
	return resp, nil
}

// UpdateIdentity updates user identity fields (full_name, avatar_url)
func (s *UserServices) UpdateIdentity(ctx context.Context, userID string, req *requests.UpdateIdentityRequest) error {
	if req.FullName != "" {
		parts := strings.SplitN(req.FullName, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}
		err := s.userRepo.UpdateUser(ctx, userID, &requests.UpdateProfileRequest{
			FirstName: firstName,
			LastName:  lastName,
		})
		if err != nil {
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
func (s *UserServices) GetPreferences(ctx context.Context, userID, orgID string) (*models.UserPreferencesResponse, error) {
	prefs, err := s.profileRepo.GetUserPreferences(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	return prefs.ToResponse(), nil
}

// UpdatePreferences updates user preferences
func (s *UserServices) UpdatePreferences(ctx context.Context, userID, orgID string, req *requests.UpdatePreferencesRequest) error {
	prefs := &models.UserPreferences{
		UserID:               userID,
		Theme:                req.Theme,
		Timezone:             req.Timezone,
		Language:             req.Language,
		DashboardDefaultView: req.DashboardDefaultView,
		TablePageSize:        req.TablePageSize,
		AutoRefreshInterval:  req.AutoRefreshInterval,
	}
	return s.profileRepo.UpsertUserPreferences(ctx, userID, orgID, prefs)
}

// GetNotifications returns user notification settings for /api/v1/profile/notifications
func (s *UserServices) GetNotifications(ctx context.Context, userID, orgID string) (*models.UserNotificationsResponse, error) {
	notifs, err := s.profileRepo.GetUserNotifications(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	return notifs.ToResponse(), nil
}

// UpdateNotifications updates user notification settings
func (s *UserServices) UpdateNotifications(ctx context.Context, userID, orgID string, req *requests.UpdateNotificationsRequest) error {
	notifs := &models.UserNotifications{
		UserID:                    userID,
		EmailEnabled:              req.EmailEnabled,
		PushEnabled:               req.PushEnabled,
		SlackEnabled:              req.SlackEnabled,
		AlertSeverityThreshold:    req.AlertSeverityThreshold,
		NotifyOnNewAlert:          req.NotifyOnNewAlert,
		NotifyOnIncidentUpdate:    req.NotifyOnIncidentUpdate,
		NotifyOnPlaybookExecution: req.NotifyOnPlaybookExecution,
	}
	return s.profileRepo.UpsertUserNotifications(ctx, userID, orgID, notifs)
}

// GetSessions returns all active sessions for /api/v1/profile/sessions
func (s *UserServices) GetSessions(ctx context.Context, userID, orgID, currentSessionID string) ([]*models.UserSessionResponse, error) {
	sessions, err := s.profileRepo.GetUserSessions(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	var responses []*models.UserSessionResponse
	for _, sess := range sessions {
		responses = append(responses, sess.ToResponse(currentSessionID))
	}
	return responses, nil
}

// RevokeSession revokes a specific session for /api/v1/profile/sessions/{id}
func (s *UserServices) RevokeSession(ctx context.Context, sessionID, userID string) error {
	return s.profileRepo.RevokeUserSession(ctx, sessionID, userID)
}

// CreateSession creates a new user session
func (s *UserServices) CreateSession(ctx context.Context, userID, orgID, tokenHash, ipAddress, userAgent string) (*models.UserSession, error) {
	session := &models.UserSession{
		UserID:           userID,
		SessionTokenHash: tokenHash,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}

	if orgID != "" {
		session.OrganizationID.String = orgID
		session.OrganizationID.Valid = true
	}
	if ipAddress != "" {
		session.IPAddress.String = ipAddress
		session.IPAddress.Valid = true
	}
	if userAgent != "" {
		session.UserAgent.String = userAgent
		session.UserAgent.Valid = true
	}

	if err := s.profileRepo.CreateUserSession(ctx, session); err != nil {
		return nil, err
	}

	// Update last login time
	_ = s.profileRepo.UpdateLastLogin(ctx, userID)

	return session, nil
}

// GetActivity returns paginated user activity for /api/v1/profile/activity
func (s *UserServices) GetActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]*models.UserActivityResponse, int, error) {
	logs, total, err := s.profileRepo.GetUserActivity(ctx, userID, orgID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var responses []*models.UserActivityResponse
	for _, log := range logs {
		responses = append(responses, log.ToResponse())
	}
	return responses, total, nil
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
	log := &models.AuditLog{
		ActionType:   actionType,
		ResourceType: resourceType,
		Metadata:     []byte("{}"), // Will be overridden if metadata provided
	}

	if userID != "" {
		log.UserID.String = userID
		log.UserID.Valid = true
	}
	if orgID != "" {
		log.OrganizationID.String = orgID
		log.OrganizationID.Valid = true
	}
	if resourceID != "" {
		log.ResourceID.String = resourceID
		log.ResourceID.Valid = true
	}
	if ipAddress != "" {
		log.IPAddress.String = ipAddress
		log.IPAddress.Valid = true
	}
	if userAgent != "" {
		log.UserAgent.String = userAgent
		log.UserAgent.Valid = true
	}

	return s.profileRepo.CreateAuditLog(ctx, log)
}
