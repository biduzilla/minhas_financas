package api

import (
	"database/sql"
	"log/slog"
)

type repositories struct {
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{}
}
