package consumers

import (
	"context"
	"encoding/json"
	"log/slog"
	"ms_goal/internal/core/messaging/events"
	goaltransaction "ms_goal/internal/features/goal_transaction"
	"shared/auth/contexts"
	"shared/auth/domain"
	"uuid"
)

type TransactionEventHandler struct {
	goalService goalService
	logger      *slog.Logger
}

type goalService interface {
	Insert(
		ctx context.Context,
		model *goaltransaction.GoalTransaction,
	) error

	DeleteByTransactionId(
		ctx context.Context,
		transactionId uuid.UUID,
	) error
}

func NewTransactionEventHandler(
	goalService goalService,
	logger *slog.Logger,
) *TransactionEventHandler {
	return &TransactionEventHandler{
		goalService: goalService,
		logger:      logger,
	}
}

func (h *TransactionEventHandler) Handle(
	ctx context.Context,
	message []byte,
	topic string,
) error {
	var event events.TransactionEvent
	if err := json.Unmarshal(message, &event); err != nil {
		h.logger.Error("Failed to unmarshal goal event (poison pill)", "error", err)
		return nil
	}

	user := domain.NewAuthenticatedUser(event.UserID, "", false, nil)
	ctx = contexts.SetUser(ctx, user)

	switch topic {
	case events.TransactionGoalCreatedTopic:
		return h.handleCreated(ctx, event)
	case events.TransactionGoalDeletedTopic:
		return h.handleDeleted(ctx, event)
	default:
		h.logger.Warn("Unknown topic ", "type", topic)
		return nil
	}
}

func (h *TransactionEventHandler) handleCreated(
	ctx context.Context,
	event events.TransactionEvent,
) error {
	model := goaltransaction.GoalTransaction{
		TransactionID: event.ID,
		UserID:        event.UserID,
		GoalID:        event.GoalID,
		Amount:        event.Amount,
	}

	return h.goalService.Insert(ctx, &model)
}

func (h *TransactionEventHandler) handleDeleted(
	ctx context.Context,
	event events.TransactionEvent,
) error {
	return h.goalService.DeleteByTransactionId(ctx, event.ID)
}
