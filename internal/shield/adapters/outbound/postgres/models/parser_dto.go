package models

import (
	"encoding/json"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/domain"
	"time"

	"github.com/google/uuid"
)

type ParserDTO struct {
	ID              uuid.UUID          `db:"id"`
	OrganizationID  uuid.UUID          `db:"organization_id"`
	SourceID        *uuid.UUID         `db:"source_id"`
	Name            string             `db:"name"`
	Description     *string            `db:"description"`
	ParserType      types.ParserType   `db:"parser_type"`
	Status          types.ParserStatus `db:"status"`
	Tags            json.RawMessage    `db:"tags"`
	Logic           json.RawMessage    `db:"logic"`
	Mappings        json.RawMessage    `db:"mappings"`
	EventsParsed24h int64              `db:"events_parsed_24h"`
	ErrorRate       float64            `db:"error_rate"`
	OwnerUserID     *uuid.UUID         `db:"owner_user_id"`
	CreatedAt       time.Time          `db:"created_at"`
	UpdatedAt       time.Time          `db:"updated_at"`
	DeletedAt       *time.Time         `db:"deleted_at"`
}

func (dto *ParserDTO) ToDomain() *domain.Parser {
	if dto == nil {
		return nil
	}

	var tags []string
	if len(dto.Tags) > 0 {
		_ = json.Unmarshal(dto.Tags, &tags)
	}

	var logic map[string]interface{}
	if len(dto.Logic) > 0 {
		_ = json.Unmarshal(dto.Logic, &logic)
	}

	var mappings []map[string]interface{}
	if len(dto.Mappings) > 0 {
		_ = json.Unmarshal(dto.Mappings, &mappings)
	}

	return &domain.Parser{
		ID:              dto.ID,
		OrganizationID:  dto.OrganizationID,
		SourceID:        dto.SourceID,
		Name:            dto.Name,
		Description:     dto.Description,
		ParserType:      dto.ParserType,
		Status:          dto.Status,
		Tags:            tags,
		Logic:           logic,
		Mappings:        mappings,
		EventsParsed24h: dto.EventsParsed24h,
		ErrorRate:       dto.ErrorRate,
		OwnerUserID:     dto.OwnerUserID,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
		DeletedAt:       dto.DeletedAt,
	}
}

type ParserVersionDTO struct {
	ID             uuid.UUID       `db:"id"`
	OrganizationID uuid.UUID       `db:"organization_id"`
	ParserID       uuid.UUID       `db:"parser_id"`
	VersionNumber  int             `db:"version_number"`
	Logic          json.RawMessage `db:"logic"`
	Mappings       json.RawMessage `db:"mappings"`
	ChangedBy      *uuid.UUID      `db:"changed_by"`
	ChangeNote     *string         `db:"change_note"`
	CreatedAt      time.Time       `db:"created_at"`
}

func (dto *ParserVersionDTO) ToDomain() *domain.ParserVersion {
	if dto == nil {
		return nil
	}

	var logic map[string]interface{}
	if len(dto.Logic) > 0 {
		_ = json.Unmarshal(dto.Logic, &logic)
	}

	var mappings []map[string]interface{}
	if len(dto.Mappings) > 0 {
		_ = json.Unmarshal(dto.Mappings, &mappings)
	}

	return &domain.ParserVersion{
		ID:             dto.ID,
		OrganizationID: dto.OrganizationID,
		ParserID:       dto.ParserID,
		VersionNumber:  dto.VersionNumber,
		Logic:          logic,
		Mappings:       mappings,
		ChangedBy:      dto.ChangedBy,
		ChangeNote:     dto.ChangeNote,
		CreatedAt:      dto.CreatedAt,
	}
}
