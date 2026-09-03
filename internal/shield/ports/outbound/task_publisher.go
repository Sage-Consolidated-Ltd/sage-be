package outbound

import (
	"context"

	"sage-backend/internal/shield/ports/dto"

	"github.com/google/uuid"
)

type TaskPublisherInt interface {
	EnqueueSubmitLogFile(ctx context.Context, input dto.SubmitLogFileInput) error
	EnqueueQualityScanJob(ctx context.Context, jobID, orgID, scanID uuid.UUID) error
}
