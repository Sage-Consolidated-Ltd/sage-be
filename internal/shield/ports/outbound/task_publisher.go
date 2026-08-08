package outbound

import (
	"context"

	"sage-backend/internal/shield/ports/dto"
)

type TaskPublisherInt interface {
	EnqueueSubmitLogFile(ctx context.Context, input dto.SubmitLogFileInput) error
}
