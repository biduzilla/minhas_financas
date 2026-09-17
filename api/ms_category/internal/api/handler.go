package api

import (
	"ms_category/internal/features/category"
	"shared/auth/domain/apiError"
)

type handlers struct {
	category *category.CategoryHandler
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{
		category: category.NewHandler(services.category, errHandler),
	}
}
