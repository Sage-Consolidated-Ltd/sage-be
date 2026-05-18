package services

import (
	"context"
	"errors"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/users/models"
	"sage-backend/internal/users/requests"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, req *requests.CreateUserRequest, hash string) error {
	return nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) MarkEmailVerified(ctx context.Context, email string) error {
	return nil
}

func (m *mockUserRepo) CreateUserWithOrganization(ctx context.Context, req *requests.CreateUserRequest, hash string) error {
	return nil
}

func (m *mockUserRepo) GetUserOrganizations(ctx context.Context, userId string) (*[]models.Organization, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Organization), args.Error(1)
}

func (m *mockUserRepo) Enable2FA(ctx context.Context, secret string, userID string) error {
	return nil
}

func (m *mockUserRepo) GetTOTPSecret(ctx context.Context, userID string) (string, error) {
	return "", nil
}

func (m *mockUserRepo) UpdateUserPassword(ctx context.Context, email string, hash string) error {
	return nil
}

func (m *mockUserRepo) OnboardUserWithTransaction(ctx context.Context, req *requests.OnboardingRequest, hash string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, id string, req *requests.UpdateProfileRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

type mockProfileRepo struct{}

func (m *mockProfileRepo) GetUserPreferences(ctx context.Context, userID, orgID string) (*models.UserPreferences, error) {
	return nil, nil
}
func (m *mockProfileRepo) UpsertUserPreferences(ctx context.Context, userID, orgID string, prefs *models.UserPreferences) error {
	return nil
}
func (m *mockProfileRepo) GetUserNotifications(ctx context.Context, userID, orgID string) (*models.UserNotifications, error) {
	return nil, nil
}
func (m *mockProfileRepo) UpsertUserNotifications(ctx context.Context, userID, orgID string, notifs *models.UserNotifications) error {
	return nil
}
func (m *mockProfileRepo) GetUserSessions(ctx context.Context, userID, orgID string) ([]models.UserSession, error) {
	return nil, nil
}
func (m *mockProfileRepo) GetUserSessionByID(ctx context.Context, sessionID string) (*models.UserSession, error) {
	return nil, nil
}
func (m *mockProfileRepo) CreateUserSession(ctx context.Context, session *models.UserSession) error {
	return nil
}
func (m *mockProfileRepo) RevokeUserSession(ctx context.Context, sessionID, userID string) error {
	return nil
}
func (m *mockProfileRepo) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	return nil
}
func (m *mockProfileRepo) GetUserActivity(ctx context.Context, userID, orgID string, page, pageSize int) ([]models.AuditLog, int, error) {
	return nil, 0, nil
}
func (m *mockProfileRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return nil
}
func (m *mockProfileRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	return nil
}
func (m *mockProfileRepo) UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error {
	return nil
}

func TestUserServices_GetProfile_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	profileRepo := &mockProfileRepo{}
	svc := NewUsersServices(mockRepo, profileRepo, &redis.Client{}, nil)

	mockUser := &models.User{
		ID:        "user-123",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Role:      "owner",
		CreatedAt: time.Now(),
	}

	mockOrgs := &[]models.Organization{
		{
			ID:        "org-123",
			Name:      "Acme Corp",
			OwnerID:   "user-123",
			CreatedAt: time.Now(),
		},
	}

	mockRepo.On("GetUserByID", mock.Anything, "user-123").Return(mockUser, nil)
	mockRepo.On("GetUserOrganizations", mock.Anything, "user-123").Return(mockOrgs, nil)

	profile, err := svc.GetProfile(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "user-123", profile.ID)
	assert.Equal(t, "John", profile.FirstName)
	assert.Equal(t, "Doe", profile.LastName)
	assert.Equal(t, "john@example.com", profile.Email)
	assert.Equal(t, "owner", profile.Role)
	assert.Len(t, profile.Organization, 1)

	mockRepo.AssertExpectations(t)
}

func TestUserServices_GetProfile_UserNotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	profileRepo := &mockProfileRepo{}
	svc := NewUsersServices(mockRepo, profileRepo, &redis.Client{}, nil)

	mockRepo.On("GetUserByID", mock.Anything, "user-123").Return(nil, apperrors.NotFoundError("USER NOT FOUND"))

	profile, err := svc.GetProfile(context.Background(), "user-123")

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "USER NOT FOUND")

	mockRepo.AssertExpectations(t)
}

func TestUserServices_GetProfile_OrgsError(t *testing.T) {
	mockRepo := new(mockUserRepo)
	profileRepo := &mockProfileRepo{}
	svc := NewUsersServices(mockRepo, profileRepo, &redis.Client{}, nil)

	mockUser := &models.User{
		ID:        "user-123",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Role:      "owner",
		CreatedAt: time.Now(),
	}

	mockRepo.On("GetUserByID", mock.Anything, "user-123").Return(mockUser, nil)
	mockRepo.On("GetUserOrganizations", mock.Anything, "user-123").Return(nil, errors.New("db error"))

	profile, err := svc.GetProfile(context.Background(), "user-123")

	assert.Error(t, err)
	assert.Nil(t, profile)

	mockRepo.AssertExpectations(t)
}

func TestUserServices_UpdateProfile_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	profileRepo := &mockProfileRepo{}
	svc := NewUsersServices(mockRepo, profileRepo, &redis.Client{}, nil)

	req := &requests.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		TimeZone:  strPtr("America/New_York"),
	}

	mockRepo.On("UpdateUser", mock.Anything, "user-123", req).Return(nil)

	err := svc.UpdateProfile(context.Background(), "user-123", req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserServices_UpdateProfile_Error(t *testing.T) {
	mockRepo := new(mockUserRepo)
	profileRepo := &mockProfileRepo{}
	svc := NewUsersServices(mockRepo, profileRepo, &redis.Client{}, nil)

	req := &requests.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}

	mockRepo.On("UpdateUser", mock.Anything, "user-123", req).Return(errors.New("update failed"))

	err := svc.UpdateProfile(context.Background(), "user-123", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	mockRepo.AssertExpectations(t)
}

func strPtr(s string) *string {
	return &s
}
