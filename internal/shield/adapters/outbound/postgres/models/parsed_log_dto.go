package models

import (
	"database/sql"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shield/domain"

	"github.com/google/uuid"
)

type ParsedLogDTO struct {
	ID           uuid.UUID    `db:"id"`
	DataSourceID uuid.UUID    `db:"data_source_id"`
	FileID       uuid.UUID    `db:"file_id"`
	Timestamp    sql.NullTime `db:"timestamp"`
	Level        string       `db:"level"`
	Message      string       `db:"message"`
	RawJSON      db.JSONMap   `db:"raw_json"`
}

func (dto *ParsedLogDTO) ToDomain() domain.ParsedLog {
	return domain.ParsedLog{
		ID:           dto.ID,
		DataSourceID: dto.DataSourceID,
		FileID:       dto.FileID,
		Timestamp:    dto.Timestamp,
		Level:        dto.Level,
		Message:      dto.Message,
		RawJSON:      dto.RawJSON,
	}
}
