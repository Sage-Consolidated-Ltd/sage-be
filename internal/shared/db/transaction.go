package db

import (
	"context"
)

type TransactionManager struct {
	db *DB
}

func NewTransactionManager(database *DB) *TransactionManager {
	return &TransactionManager{
		db: database,
	}
}

func (m *TransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	tx, err := m.db.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	ctx = WithExecutor(ctx, tx)

	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
