package api

import (
	"ms_transaction/internal/features/transactions"
	"shared/auth/domain/apiError"
)

type handlers struct {
	transaction *transactions.TransactionHandler
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{
		transaction: transactions.NewHandler(services.transaction, errHandler),
	}
}
