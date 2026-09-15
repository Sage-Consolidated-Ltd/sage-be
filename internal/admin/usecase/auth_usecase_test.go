package usecase_test

import (
	"context"
	"testing"
	"time"

	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/admin/usecase"
	"sage-backend/internal/admin/usecase/dto"
	"sage-backend/internal/shared/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAdminRepo struct {
	mock.Mock
}

func (m *mockAdminRepo) GetByID(ctx context.Context, id string) (*domain.Admin, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Admin), args.Error(1)
}

func (m *mockAdminRepo) GetByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Admin), args.Error(1)
}

func (m *mockAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	args := m.Called(ctx, admin)
	return args.Error(0)
}

func (m *mockAdminRepo) Update(ctx context.Context, admin *domain.Admin) error {
	args := m.Called(ctx, admin)
	return args.Error(0)
}

type mockOTPStore struct {
	mock.Mock
}

func (m *mockOTPStore) SaveOTP(ctx context.Context, data outbound.OTPPendingData, ttl time.Duration) error {
	args := m.Called(ctx, data, ttl)
	return args.Error(0)
}

func (m *mockOTPStore) VerifyAndConsumeOTP(ctx context.Context, adminID string, otp string) (*outbound.OTPPendingData, error) {
	args := m.Called(ctx, adminID, otp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.OTPPendingData), args.Error(1)
}

func (m *mockOTPStore) ClearOTP(ctx context.Context, adminID string) error {
	args := m.Called(ctx, adminID)
	return args.Error(0)
}

type mockSessionStore struct {
	mock.Mock
}

func (m *mockSessionStore) CreateSession(ctx context.Context, session *outbound.AdminSession, ttl time.Duration) error {
	args := m.Called(ctx, session, ttl)
	return args.Error(0)
}

func (m *mockSessionStore) GetSession(ctx context.Context, sessionID string) (*outbound.AdminSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*outbound.AdminSession), args.Error(1)
}

func (m *mockSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

type mockEmailService struct {
	mock.Mock
}

func (m *mockEmailService) SendAdminOTP(ctx context.Context, toEmail string, otpCode string) error {
	args := m.Called(ctx, toEmail, otpCode)
	return args.Error(0)
}

func TestAuthUseCase_LockoutStateMachine(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	rawPassword := "SecurePassword123!"
	passwordHash, err := utils.HashPassword(rawPassword)
	require.NoError(t, err)

	validAdminEmail, err := domain.NewEmail("admin@sage.com")
	require.NoError(t, err)

	t.Run("4 failed attempts keeps account active and returns ErrInvalidCredentials", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		admin := domain.ReconstituteAdmin(
			"admin-1",
			validAdminEmail,
			passwordHash,
			domain.RoleSuperAdmin,
			domain.StatusActive,
			3, // already 3 failed attempts
			nil,
			nil,
			now,
			now,
		)

		adminRepo.On("GetByEmail", ctx, "admin@sage.com").Return(admin, nil)
		adminRepo.On("Update", ctx, mock.Anything).Return(nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		result, err := uc.Login(ctx, dto.AdminLoginInput{
			Email:      "admin@sage.com",
			Password:   "wrong-password",
			RememberMe: false,
			ClientIP:   "10.0.0.1",
		})

		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.Nil(t, result)
		assert.Equal(t, 4, admin.FailedAttempts())
		assert.Equal(t, domain.StatusActive, admin.Status())
	})

	t.Run("5th failed attempt triggers temporary lockout for 30 minutes", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		admin := domain.ReconstituteAdmin(
			"admin-1",
			validAdminEmail,
			passwordHash,
			domain.RoleSuperAdmin,
			domain.StatusActive,
			4, // currently 4 failed attempts
			nil,
			nil,
			now,
			now,
		)

		adminRepo.On("GetByEmail", ctx, "admin@sage.com").Return(admin, nil)
		adminRepo.On("Update", ctx, mock.Anything).Return(nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		result, err := uc.Login(ctx, dto.AdminLoginInput{
			Email:      "admin@sage.com",
			Password:   "wrong-password",
			RememberMe: false,
			ClientIP:   "10.0.0.1",
		})

		assert.ErrorIs(t, err, domain.ErrAccountTemporarilyLocked)
		assert.Nil(t, result)
		assert.Equal(t, 5, admin.FailedAttempts())
		assert.Equal(t, domain.StatusTemporarilyLocked, admin.Status())
		require.NotNil(t, admin.LockedUntil())
		assert.True(t, admin.LockedUntil().After(now.Add(29*time.Minute)))
	})

	t.Run("Login attempt while temporarily locked is blocked immediately", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		futureLock := now.Add(20 * time.Minute)
		admin := domain.ReconstituteAdmin(
			"admin-1",
			validAdminEmail,
			passwordHash,
			domain.RoleSuperAdmin,
			domain.StatusTemporarilyLocked,
			5,
			&futureLock,
			nil,
			now,
			now,
		)

		adminRepo.On("GetByEmail", ctx, "admin@sage.com").Return(admin, nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		result, err := uc.Login(ctx, dto.AdminLoginInput{
			Email:      "admin@sage.com",
			Password:   rawPassword, // even with correct password!
			RememberMe: false,
			ClientIP:   "10.0.0.1",
		})

		assert.ErrorIs(t, err, domain.ErrAccountTemporarilyLocked)
		assert.Nil(t, result)
	})

	t.Run("10th failed attempt triggers permanent lockout", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		pastLock := now.Add(-1 * time.Minute) // temporary lock expired
		admin := domain.ReconstituteAdmin(
			"admin-1",
			validAdminEmail,
			passwordHash,
			domain.RoleSuperAdmin,
			domain.StatusTemporarilyLocked,
			9, // 9 failed attempts
			&pastLock,
			nil,
			now,
			now,
		)

		adminRepo.On("GetByEmail", ctx, "admin@sage.com").Return(admin, nil)
		adminRepo.On("Update", ctx, mock.Anything).Return(nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		result, err := uc.Login(ctx, dto.AdminLoginInput{
			Email:      "admin@sage.com",
			Password:   "wrong-password",
			RememberMe: false,
			ClientIP:   "10.0.0.1",
		})

		assert.ErrorIs(t, err, domain.ErrAccountPermanentlyLocked)
		assert.Nil(t, result)
		assert.Equal(t, 10, admin.FailedAttempts())
		assert.Equal(t, domain.StatusPermanentlyLocked, admin.Status())
	})

	t.Run("Rejects invalid email format without valid pattern", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		result, err := uc.Login(ctx, dto.AdminLoginInput{
			Email:      "admin@localhost",
			Password:   "password123!",
			RememberMe: false,
			ClientIP:   "10.0.0.1",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAuthUseCase_SuccessfulLoginFlow(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	rawPassword := "SuperAdminSecret99!"
	passwordHash, err := utils.HashPassword(rawPassword)
	require.NoError(t, err)

	superAdminEmail, err := domain.NewEmail("super@sage.com")
	require.NoError(t, err)

	adminRepo := new(mockAdminRepo)
	otpStore := new(mockOTPStore)
	sessionStore := new(mockSessionStore)
	emailService := new(mockEmailService)

	admin := domain.ReconstituteAdmin(
		"admin-uuid-1",
		superAdminEmail,
		passwordHash,
		domain.RoleSuperAdmin,
		domain.StatusActive,
		2,
		nil,
		nil,
		now,
		now,
	)

	adminRepo.On("GetByEmail", ctx, "super@sage.com").Return(admin, nil)
	otpStore.On("SaveOTP", ctx, mock.MatchedBy(func(data outbound.OTPPendingData) bool {
		return data.AdminID == "admin-uuid-1" && len(data.OTP) == 6 && data.RememberMe == true && data.ClientIP == "192.168.1.100"
	}), 10*time.Minute).Return(nil)
	emailService.On("SendAdminOTP", ctx, "super@sage.com", mock.AnythingOfType("string")).Return(nil)

	uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)

	// Step 1: Login challenge
	loginRes, err := uc.Login(ctx, dto.AdminLoginInput{
		Email:      "super@sage.com",
		Password:   rawPassword,
		RememberMe: true,
		ClientIP:   "192.168.1.100",
	})
	require.NoError(t, err)
	require.NotNil(t, loginRes)
	assert.Equal(t, "admin-uuid-1", loginRes.AdminID)

	// Step 2: Verify OTP
	adminRepo.On("GetByID", ctx, "admin-uuid-1").Return(admin, nil)
	otpStore.On("VerifyAndConsumeOTP", ctx, "admin-uuid-1", "123456").Return(&outbound.OTPPendingData{
		AdminID:    "admin-uuid-1",
		OTP:        "123456",
		RememberMe: true,
		ClientIP:   "192.168.1.100",
	}, nil)
	adminRepo.On("Update", ctx, admin).Return(nil)
	sessionStore.On("CreateSession", ctx, mock.MatchedBy(func(sess *outbound.AdminSession) bool {
		return sess.AdminID == "admin-uuid-1" && sess.BoundIP == "192.168.1.100" && sess.Role == "super_admin"
	}), 14*24*time.Hour).Return(nil)

	verifyRes, err := uc.VerifyOTP(ctx, dto.VerifyOTPInput{
		AdminID:  "admin-uuid-1",
		OTP:      "123456",
		ClientIP: "192.168.1.100",
	})
	require.NoError(t, err)
	require.NotNil(t, verifyRes)
	assert.NotEmpty(t, verifyRes.SessionID)
	assert.Equal(t, "super@sage.com", verifyRes.Admin.Email)
	assert.Equal(t, 0, admin.FailedAttempts())
	assert.NotNil(t, admin.LastLoginAt())
}

func TestAuthUseCase_SessionIPAnomalyDetection(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	adminRepo := new(mockAdminRepo)
	otpStore := new(mockOTPStore)
	sessionStore := new(mockSessionStore)
	emailService := new(mockEmailService)

	session := &outbound.AdminSession{
		SessionID: "session-abc-123",
		AdminID:   "admin-1",
		Role:      "super_admin",
		Email:     "admin@sage.com",
		BoundIP:   "192.168.1.50",
		CreatedAt: now,
		ExpiresAt: now.Add(12 * time.Hour),
	}

	t.Run("Matching client IP validates successfully", func(t *testing.T) {
		sessionStore.On("GetSession", ctx, "session-abc-123").Return(session, nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		sess, err := uc.ValidateSession(ctx, "session-abc-123", "192.168.1.50")
		require.NoError(t, err)
		assert.Equal(t, "admin-1", sess.AdminID)
	})

	t.Run("Different client IP triggers immediate session revocation and ErrIPMismatch", func(t *testing.T) {
		sessionStore.On("GetSession", ctx, "session-abc-123").Return(session, nil)
		sessionStore.On("DeleteSession", ctx, "session-abc-123").Return(nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		sess, err := uc.ValidateSession(ctx, "session-abc-123", "203.0.113.99") // hijacked IP!

		assert.ErrorIs(t, err, domain.ErrIPMismatch)
		assert.Nil(t, sess)
		sessionStore.AssertCalled(t, "DeleteSession", ctx, "session-abc-123")
	})
}

func TestAuthUseCase_BootstrapSuperAdmin(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully bootstraps super admin on fresh database", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		email := "superadmin@sage.com"
		password := "SuperSecretPassword123!"

		adminRepo.On("GetByEmail", ctx, email).Return(nil, domain.ErrAdminNotFound)
		adminRepo.On("Create", ctx, mock.MatchedBy(func(a *domain.Admin) bool {
			return a.Email().String() == email &&
				a.Role() == domain.RoleSuperAdmin &&
				a.Status() == domain.StatusActive &&
				utils.CompareHashAndPassword(password, a.PasswordHash())
		})).Return(nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		err := uc.BootstrapSuperAdmin(ctx, email, password)

		require.NoError(t, err)
		adminRepo.AssertExpectations(t)
	})

	t.Run("Idempotent when super admin already exists", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		email := "superadmin@sage.com"
		password := "SuperSecretPassword123!"

		emailVO, err := domain.NewEmail(email)
		require.NoError(t, err)
		existingAdmin := domain.NewAdmin("admin-existing-id", emailVO, "existing-hash", domain.RoleSuperAdmin, time.Now().UTC())

		adminRepo.On("GetByEmail", ctx, email).Return(existingAdmin, nil)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		err = uc.BootstrapSuperAdmin(ctx, email, password)

		require.NoError(t, err)
		adminRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("Rejects invalid email format", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		err := uc.BootstrapSuperAdmin(ctx, "invalid-email-pattern", "SuperSecretPassword123!")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid super admin email")
	})

	t.Run("Rejects short password (< 8 characters)", func(t *testing.T) {
		adminRepo := new(mockAdminRepo)
		otpStore := new(mockOTPStore)
		sessionStore := new(mockSessionStore)
		emailService := new(mockEmailService)

		uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
		err := uc.BootstrapSuperAdmin(ctx, "superadmin@sage.com", "short")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at least 8 characters")
	})
}

