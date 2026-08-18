package api

import (
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/features/category"
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
