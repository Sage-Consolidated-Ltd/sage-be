package outbound

import (
	"context"

	"sage-backend/internal/shield/domain"
	"sage-backend/internal/shield/ports/dto"
)

type ThreatDetectorInt interface {
	SubmitLogFileForAnalysis(
		ctx context.Context,
		input domain.SubmitLogFileForAnalysis,
	) (*dto.SubmitLogFileResult, error)
}
