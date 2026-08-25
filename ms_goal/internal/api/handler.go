package api

import (
	"ms_goal/internal/core/domain/apiError"
)

type handlers struct {
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{}
}
