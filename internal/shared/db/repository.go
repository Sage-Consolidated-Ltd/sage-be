package db

import "context"

type Repository struct {
	db *DB
}

func NewRepository(database *DB) Repository {
	return Repository{
		db: database,
	}
}

func (r Repository) Executor(ctx context.Context) Executor {
	return r.db.Executor(ctx)
}
