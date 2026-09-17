package api

import (
	"database/sql"
	"log/slog"
	"ms_transaction/internal/features/transactions"
)

type repositories struct {
	transaction *transactions.TransactionRepository
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		transaction: transactions.NewRepository(db, logger),
	}
}
