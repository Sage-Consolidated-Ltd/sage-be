package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUploadService struct {
	mock.Mock
}

func (m *mockUploadService) GetUploadURL(
	ctx context.Context,
	rc middlewares.RequestContext,
	orgID uuid.UUID,
	filename string,
	contentType string,
	sizeBytes int64,
) (*dto.PresignUploadResponse, error) {
	args := m.Called(ctx, rc, orgID, filename, contentType, sizeBytes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PresignUploadResponse), args.Error(1)
}

func (m *mockUploadService) ValidateUploadComplete(
	ctx context.Context,
	rc *middlewares.RequestContext,
	req *dto.UploadCompleteRequest,
) (*domain.LogFile, error) {
	args := m.Called(ctx, rc, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LogFile), args.Error(1)
}

func TestUploadHandler_UploadLog(t *testing.T) {
	mockSvc := new(mockUploadService)
	handler := NewUploadHandler(mockSvc)

	orgID := uuid.New()
	userID := uuid.New().String()

	setupApp := func() *fiber.App {
		app := fiber.New()
		app.Post("/upload/presign", func(c *fiber.Ctx) error {
			c.Locals("orgID", orgID)
			c.Locals("userID", userID)
			return handler.UploadLog(c)
		})
		return app
	}

	t.Run("successful presign", func(t *testing.T) {
		app := setupApp()
		reqBody := dto.UploadLogRequest{
			Filename:    "windows_security.csv",
			ContentType: "text/csv",
			Size:        1024,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockSvc.On("GetUploadURL",
			mock.Anything,
			mock.Anything,
			orgID,
			"windows_security.csv",
			"text/csv",
			int64(1024),
		).Return(&dto.PresignUploadResponse{
			Key:       "uploads/pending/csv/test.csv",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}, nil).Once()

		req := httptest.NewRequest("POST", "/upload/presign", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid request validation", func(t *testing.T) {
		app := setupApp()
		reqBody := dto.UploadLogRequest{
			Filename: "", // missing required filename
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/upload/presign", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	})
}
