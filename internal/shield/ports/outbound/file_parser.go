package outbound

import "sage-backend/internal/shield/domain"

type LogFileParserInt interface {
	Parse(fileBytes []byte) ([]domain.ParsedLog, error)
}
