package api

import (
	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/features/goal"
	goaltransaction "ms_goal/internal/features/goal_transaction"
)

type handlers struct {
	goal            *goal.GoalHandler
	goalTransaction *goaltransaction.GoalTransactionHandler
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{
		goal:            goal.NewHandler(services.goal, errHandler),
		goalTransaction: goaltransaction.NewHandler(services.goalTransaction, errHandler),
	}
}
