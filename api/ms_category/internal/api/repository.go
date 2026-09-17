package api

import (
	"database/sql"
	"log/slog"
	"ms_category/internal/features/category"
)

type repositories struct {
	category *category.CategoryRepository
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		category: category.NewRepository(db, logger),
	}
}
