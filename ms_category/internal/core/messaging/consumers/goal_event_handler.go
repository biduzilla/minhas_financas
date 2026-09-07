package consumers

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"ms_category/internal/core/contexts"
	"ms_category/internal/core/domain"
	"ms_category/internal/core/messaging/events"
	"ms_category/internal/features/category"
	"uuid"
)

type GoalEventHandler struct {
	categoryService categoryService
	logger          *slog.Logger
}

type categoryService interface {
	Insert(
		ctx context.Context,
		model *category.Category,
	) error

	DeleteByGoalId(
		ctx context.Context,
		id uuid.UUID,
	) error
}

func NewGoalEventHandler(categoryService categoryService, logger *slog.Logger) *GoalEventHandler {
	return &GoalEventHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

func (h *GoalEventHandler) Handle(ctx context.Context, message []byte, topic string) error {
	var event events.GoalEvent
	if err := json.Unmarshal(message, &event); err != nil {
		h.logger.Error("Failed to unmarshal goal event (poison pill)", "error", err)
		return nil
	}

	user := domain.NewAuthenticatedUser(event.UserID, "", false, nil)
	ctx = contexts.SetUser(ctx, user)

	switch topic {
	case events.GoalCreatedTopic:
		return h.handleCreated(ctx, event)
	case events.GoalDeletedTopic:
		return h.handleDeleted(ctx, event)
	default:
		h.logger.Warn("Unknown topic ", "type", topic)
		return nil
	}
}

func (h *GoalEventHandler) handleCreated(ctx context.Context, event events.GoalEvent) error {
	model := category.Category{
		GoalID: &event.ID,
		Name:   event.Name,
	}
	return h.categoryService.Insert(ctx, &model)
}

func (h *GoalEventHandler) handleDeleted(ctx context.Context, event events.GoalEvent) error {
	return h.categoryService.DeleteByGoalId(ctx, event.ID)
}
