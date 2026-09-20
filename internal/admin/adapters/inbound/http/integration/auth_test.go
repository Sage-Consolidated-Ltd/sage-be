package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adminHttp "sage-backend/internal/admin/adapters/inbound/http"
	"sage-backend/internal/admin/adapters/inbound/http/dto"
	"sage-backend/internal/admin/domain"
	"sage-backend/internal/admin/ports/outbound"
	"sage-backend/internal/admin/usecase"
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
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

func performTestRequest(app *fiber.App, method, path string, body any, cookie *http.Cookie) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}

	return app.Test(req, -1)
}

func TestAdminAuthHTTP_Flow(t *testing.T) {
	now := time.Now().UTC()
	password := "PlatformAdminPass99!"
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	adminRepo := new(mockAdminRepo)
	otpStore := new(mockOTPStore)
	sessionStore := new(mockSessionStore)
	emailService := new(mockEmailService)

	cfg := &config.BaseConfig{APP_ENV: "testing"}
	uc := usecase.NewAuthUseCase(adminRepo, otpStore, sessionStore, emailService)
	handler := adminHttp.NewAuthHandler(uc, cfg)
	middleware := adminHttp.NewAdminAuthMiddleware(uc)

	app := fiber.New()
	adminHttp.SetUpAuthRouter(app, handler, middleware)

	adminEmail, err := domain.NewEmail("ops@sageconsolidated.com")
	require.NoError(t, err)

	admin := domain.ReconstituteAdmin(
		"b91d7637-2361-460d-a77b-607519bfb82a",
		adminEmail,
		hash,
		domain.RoleSuperAdmin,
		domain.StatusActive,
		0,
		nil,
		nil,
		now,
		now,
	)

	// Step 1: Test Login
	t.Run("POST /admin/auth/login initiates OTP challenge", func(t *testing.T) {
		adminRepo.On("GetByEmail", mock.Anything, "ops@sageconsolidated.com").Return(admin, nil).Once()
		otpStore.On("SaveOTP", mock.Anything, mock.Anything, 10*time.Minute).Return(nil).Once()
		emailService.On("SendAdminOTP", mock.Anything, "ops@sageconsolidated.com", mock.Anything).Return(nil).Once()

		reqBody := dto.AdminLoginRequest{
			Email:      "ops@sageconsolidated.com",
			Password:   password,
			RememberMe: true,
		}

		resp, err := performTestRequest(app, http.MethodPost, "/admin/auth/login", reqBody, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.Response
		err = json.NewDecoder(resp.Body).Decode(&res)
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	// Step 2: Test Verify OTP
	t.Run("POST /admin/auth/verify-otp sets session cookie and returns admin data", func(t *testing.T) {
		adminRepo.On("GetByID", mock.Anything, "b91d7637-2361-460d-a77b-607519bfb82a").Return(admin, nil).Once()
		otpStore.On("VerifyAndConsumeOTP", mock.Anything, "b91d7637-2361-460d-a77b-607519bfb82a", "654321").Return(&outbound.OTPPendingData{
			AdminID:    "b91d7637-2361-460d-a77b-607519bfb82a",
			OTP:        "654321",
			RememberMe: true,
			ClientIP:   "0.0.0.0",
		}, nil).Once()
		adminRepo.On("Update", mock.Anything, admin).Return(nil).Once()
		sessionStore.On("CreateSession", mock.Anything, mock.Anything, 14*24*time.Hour).Return(nil).Once()

		reqBody := dto.VerifyOTPRequest{
			AdminID: "b91d7637-2361-460d-a77b-607519bfb82a",
			OTP:     "654321",
		}

		resp, err := performTestRequest(app, http.MethodPost, "/admin/auth/verify-otp", reqBody, nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.Response
		err = json.NewDecoder(resp.Body).Decode(&res)
		require.NoError(t, err)
		assert.True(t, res.Success)

		cookies := resp.Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == adminHttp.AdminSessionCookie {
				sessionCookie = c
				break
			}
		}
		assert.NotNil(t, sessionCookie, "admin_session cookie must be set")
	})

	// Step 3: Test Authenticated Me Endpoint
	t.Run("GET /admin/auth/me returns authenticated admin profile", func(t *testing.T) {
		session := &outbound.AdminSession{
			SessionID: "sess-test-123",
			AdminID:   "b91d7637-2361-460d-a77b-607519bfb82a",
			Role:      "super_admin",
			Email:     "ops@sageconsolidated.com",
			BoundIP:   "0.0.0.0",
			CreatedAt: now,
			ExpiresAt: now.Add(12 * time.Hour),
		}

		sessionStore.On("GetSession", mock.Anything, "sess-test-123").Return(session, nil).Once()
		adminRepo.On("GetByID", mock.Anything, "b91d7637-2361-460d-a77b-607519bfb82a").Return(admin, nil).Once()

		cookie := &http.Cookie{Name: adminHttp.AdminSessionCookie, Value: "sess-test-123"}
		resp, err := performTestRequest(app, http.MethodGet, "/admin/auth/me", nil, cookie)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.Response
		err = json.NewDecoder(resp.Body).Decode(&res)
		require.NoError(t, err)
		assert.True(t, res.Success)
	})

	// Step 4: Test Logout
	t.Run("POST /admin/auth/logout invalidates session and clears cookie", func(t *testing.T) {
		session := &outbound.AdminSession{
			SessionID: "sess-test-123",
			AdminID:   "b91d7637-2361-460d-a77b-607519bfb82a",
			Role:      "super_admin",
			Email:     "ops@sageconsolidated.com",
			BoundIP:   "0.0.0.0",
			CreatedAt: now,
			ExpiresAt: now.Add(12 * time.Hour),
		}

		sessionStore.On("GetSession", mock.Anything, "sess-test-123").Return(session, nil).Once()
		sessionStore.On("DeleteSession", mock.Anything, "sess-test-123").Return(nil).Once()

		cookie := &http.Cookie{Name: adminHttp.AdminSessionCookie, Value: "sess-test-123"}
		resp, err := performTestRequest(app, http.MethodPost, "/admin/auth/logout", nil, cookie)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
