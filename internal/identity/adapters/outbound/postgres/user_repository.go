package postgres

import (
	"sage-backend/internal/shared/db"
	"sage-backend/internal/identity/ports/outbound"
)

type UserRepository struct {
	db.Repository
}

func NewUserRepository(database *db.DB) outbound.UserRepository {
	return &UserRepository{
		Repository: db.NewRepository(database),
	}
}

var _ outbound.UserRepository = (*UserRepository)(nil)
