package repositories

import (
	"context"
	"database/sql"
	"time"

	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/types"
	"sage-backend/internal/shield/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ParserRepositoryInt interface {
	CreateParser(ctx context.Context, parser *models.Parser) error
	UpdateParser(ctx context.Context, parser *models.Parser) error
	GetParserByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error)
	ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.Parser, int, error)
	EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	CreateParserVersion(ctx context.Context, version *models.ParserVersion) error
	GetParserVersions(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID) ([]*models.ParserVersion, error)
	CreateTestRun(ctx context.Context, run *models.ParserTestRun) error
	GetParserSummary(ctx context.Context, orgID uuid.UUID) (total, active int, errorRate float64, lastUpdated *time.Time, err error)
	ImportParser(ctx context.Context, parser *models.Parser) error
	IncrementParserMetrics(ctx context.Context, parserID uuid.UUID, eventsParsed int64, errorRate float64) error
}

type ParserRepository struct {
	db *sqlx.DB
}

func NewParserRepository(db *sqlx.DB) ParserRepositoryInt {
	return &ParserRepository{db: db}
}

const (
	CREATE_PARSER = `
		INSERT INTO parsers (
			organization_id, source_id, name, description, parser_type, status,
			tags, logic, mappings, owner_user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at, updated_at
	`
	UPDATE_PARSER = `
		UPDATE parsers SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			parser_type = COALESCE($5, parser_type),
			status = COALESCE($6, status),
			tags = COALESCE($7, tags),
			logic = COALESCE($8, logic),
			mappings = COALESCE($9, mappings),
			owner_user_id = COALESCE($10, owner_user_id),
			updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`
	GET_PARSER   = `SELECT * FROM parsers WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	LIST_PARSERS = `
		SELECT * FROM parsers
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND ($2::varchar IS NULL OR status = $2)
			AND ($3::varchar IS NULL OR parser_type = $3)
			AND ($4::varchar IS NULL OR name ILIKE '%' || $4 || '%')
		ORDER BY updated_at DESC
		LIMIT $5 OFFSET $6
	`
	COUNT_PARSERS = `
		SELECT COUNT(*) FROM parsers
		WHERE organization_id = $1 AND deleted_at IS NULL
			AND ($2::varchar IS NULL OR status = $2)
			AND ($3::varchar IS NULL OR parser_type = $3)
			AND ($4::varchar IS NULL OR name ILIKE '%' || $4 || '%')
	`
	ENABLE_PARSER         = `UPDATE parsers SET status = $2, updated_at = NOW() WHERE id = $1 AND organization_id = $3`
	DISABLE_PARSER        = `UPDATE parsers SET status = $2, updated_at = NOW() WHERE id = $1 AND organization_id = $3`
	CREATE_PARSER_VERSION = `
		INSERT INTO parser_versions (organization_id, parser_id, version_number, logic, mappings, changed_by, change_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at
	`
	GET_PARSER_VERSIONS = `SELECT * FROM parser_versions WHERE parser_id = $1 AND organization_id = $2 ORDER BY version_number DESC`
	CREATE_TEST_RUN     = `
		INSERT INTO parser_test_runs (
			organization_id, parser_id, sample_log, raw_payload, parsed_output, normalized_output, errors, success
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at
	`
	GET_PARSER_SUMMARY = `
		SELECT COUNT(*), 
		       COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0),
		       COALESCE(AVG(error_rate), 0),
		       MAX(updated_at)
		FROM parsers
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	IMPORT_PARSER            = `INSERT INTO parsers (organization_id, source_id, name, description, parser_type, status, tags, logic, mappings, owner_user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`
	INCREMENT_PARSER_METRICS = `UPDATE parsers SET events_parsed_24h = events_parsed_24h + $2, error_rate = $3, updated_at = NOW() WHERE id = $1`
)

func (r *ParserRepository) CreateParser(ctx context.Context, parser *models.Parser) error {
	var id uuid.UUID
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_PARSER,
		parser.OrganizationID, parser.SourceID, parser.Name, parser.Description,
		parser.ParserType, parser.Status, parser.Tags, parser.Logic, parser.Mappings, parser.OwnerUserID,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}
	parser.ID = id
	parser.CreatedAt = createdAt
	parser.UpdatedAt = updatedAt
	return nil
}

func (r *ParserRepository) UpdateParser(ctx context.Context, parser *models.Parser) error {
	result, err := r.db.ExecContext(
		ctx, UPDATE_PARSER,
		parser.ID, parser.OrganizationID, parser.Name, parser.Description,
		parser.ParserType, parser.Status, parser.Tags, parser.Logic, parser.Mappings, parser.OwnerUserID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("PARSER NOT FOUND")
	}
	return nil
}

func (r *ParserRepository) GetParserByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Parser, error) {
	var parser models.Parser
	err := r.db.GetContext(ctx, &parser, GET_PARSER, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFoundError("PARSER NOT FOUND")
		}
		return nil, err
	}
	return &parser, nil
}

func (r *ParserRepository) ListParsers(ctx context.Context, orgID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*models.Parser, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	status := ""
	if s, ok := filters["status"].(string); ok {
		status = s
	}
	parserType := ""
	if pt, ok := filters["parser_type"].(string); ok {
		parserType = pt
	}
	search := ""
	if s, ok := filters["search"].(string); ok {
		search = s
	}

	var total int
	err := r.db.GetContext(ctx, &total, COUNT_PARSERS, orgID, status, parserType, search)
	if err != nil {
		return nil, 0, err
	}

	var parsers []*models.Parser
	err = r.db.SelectContext(ctx, &parsers, LIST_PARSERS, orgID, status, parserType, search, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return parsers, total, nil
}

func (r *ParserRepository) EnableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, ENABLE_PARSER, id, types.ParserStatusActive, orgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("PARSER NOT FOUND")
	}
	return nil
}

func (r *ParserRepository) DisableParser(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, DISABLE_PARSER, id, types.ParserStatusDisabled, orgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NotFoundError("PARSER NOT FOUND")
	}
	return nil
}

func (r *ParserRepository) CreateParserVersion(ctx context.Context, version *models.ParserVersion) error {
	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_PARSER_VERSION,
		version.OrganizationID, version.ParserID, version.VersionNumber,
		version.Logic, version.Mappings, version.ChangedBy, version.ChangeNote,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	version.ID = id
	version.CreatedAt = createdAt
	return nil
}

func (r *ParserRepository) GetParserVersions(ctx context.Context, parserID uuid.UUID, orgID uuid.UUID) ([]*models.ParserVersion, error) {
	var versions []*models.ParserVersion
	// Use query that also filters org
	query := GET_PARSER_VERSIONS + " AND organization_id = $2"
	err := r.db.SelectContext(ctx, &versions, query, parserID, orgID)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *ParserRepository) CreateTestRun(ctx context.Context, run *models.ParserTestRun) error {
	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(
		ctx, CREATE_TEST_RUN,
		run.OrganizationID, run.ParserID, run.SampleLog, run.RawPayload,
		run.ParsedOutput, run.NormalizedOutput, run.Errors, run.Success,
	).Scan(&id, &createdAt)
	if err != nil {
		return err
	}
	run.ID = id
	run.CreatedAt = createdAt
	return nil
}

func (r *ParserRepository) GetParserSummary(ctx context.Context, orgID uuid.UUID) (total int, active int, errorRate float64, lastUpdated *time.Time, err error) {
	row := r.db.QueryRowContext(ctx, GET_PARSER_SUMMARY, orgID)
	err = row.Scan(&total, &active, &errorRate, &lastUpdated)
	return
}

func (r *ParserRepository) ImportParser(ctx context.Context, parser *models.Parser) error {
	var id uuid.UUID
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(
		ctx, IMPORT_PARSER,
		parser.OrganizationID, parser.SourceID, parser.Name, parser.Description,
		parser.ParserType, parser.Status, parser.Tags, parser.Logic, parser.Mappings, parser.OwnerUserID,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}
	parser.ID = id
	parser.CreatedAt = createdAt
	parser.UpdatedAt = updatedAt
	return nil
}

func (r *ParserRepository) IncrementParserMetrics(ctx context.Context, parserID uuid.UUID, eventsParsed int64, errorRate float64) error {
	_, err := r.db.ExecContext(ctx, INCREMENT_PARSER_METRICS, parserID, eventsParsed, errorRate)
	return err
}
