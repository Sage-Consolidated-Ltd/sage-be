package repositories

import (
	"context"
	"database/sql"
	"errors"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/users/models"
)

type CompanyRepositoryInt interface {
	GetIndustries(ctx context.Context) (*[]models.Industry, error)
	GetIndustryByID(ctx context.Context, id string) (*models.Industry, error)
}

type CompanyRepository struct {
	db *db.DB
}

var (
	GET_INDUSTRIES=`
		SELECT id, name FROM industries ORDER BY name ASC
	`
	GET_INDUSTRY_BY_ID=`
		SELECT id, name FROM industries WHERE id = $1
	`
)

func NewCompanyRepository(db *db.DB) CompanyRepositoryInt {
	return &CompanyRepository{
		db: db,
	}
}

func (r *CompanyRepository) GetIndustries(ctx context.Context) (*[]models.Industry, error) {
	var industries []models.Industry
	err := r.db.SelectContext(ctx, &industries, GET_INDUSTRIES)
	if err != nil {
		return nil, err
	}
	return &industries, nil
}

func (r *CompanyRepository) GetIndustryByID(ctx context.Context, id string) (*models.Industry, error) {
	var industry models.Industry
	err := r.db.GetContext(ctx, &industry, GET_INDUSTRY_BY_ID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, apperrors.NotFoundError("Industry not found")
		}
		return nil, err
	}
	return &industry, nil
}