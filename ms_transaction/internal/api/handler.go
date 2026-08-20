package api

import (
	"ms_transaction/internal/core/domain/apiError"
)

type handlers struct {
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{}
}
