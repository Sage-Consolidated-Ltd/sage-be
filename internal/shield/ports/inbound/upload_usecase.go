package inbound

import (
	"context"

	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"

	"github.com/google/uuid"
)

type UploadUseCase interface {
	GetUploadURL(
		ctx context.Context,
		rc middlewares.RequestContext,
		orgID uuid.UUID,
		filename string,
		contentType string,
		sizeBytes int64,
	) (*dto.PresignUploadResponse, error)

	ValidateUploadComplete(
		ctx context.Context,
		rc *middlewares.RequestContext,
		req *dto.UploadCompleteRequest,
	) (*domain.LogFile, error)
}
