package api

import (
	"database/sql"
	"log/slog"
	"ms_goal/internal/features/goal"
	goaltransaction "ms_goal/internal/features/goal_transaction"
)

type repositories struct {
	goal            *goal.GoalRepository
	goalTransaction *goaltransaction.GoalTransactionRepository
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		goal:            goal.NewRepository(db, logger),
		goalTransaction: goaltransaction.NewRepository(db, logger),
	}
}
