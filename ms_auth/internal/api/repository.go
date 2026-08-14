package api

import (
	"database/sql"
	"log/slog"
	"ms_auth/internal/features/user"
)

type repositories struct {
	user *user.UserRepository
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		user: user.NewRepository(db, logger),
	}
}
