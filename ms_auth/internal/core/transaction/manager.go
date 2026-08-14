package transaction

import (
	"context"
	"database/sql"
	"errors"
	"ms_auth/internal/core/contexts"
)

type Manager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *manager {
	return &manager{db: db}
}

func (m *manager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := contexts.GetTx(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTx := contexts.SetTx(ctx, tx)

	fnErr := fn(ctxWithTx)
	if fnErr == nil {
		if commitErr := tx.Commit(); commitErr != nil {
			return commitErr
		}
		return nil
	}

	if rbErr := tx.Rollback(); rbErr != nil {
		return errors.Join(fnErr, rbErr)
	}

	return fnErr
}
