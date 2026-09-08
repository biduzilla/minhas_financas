package producers

import (
	"context"
	"log/slog"
	"messaging"
	"ms_goal/internal/core/messaging/events"

	"github.com/IBM/sarama"
)

const (
	GoalCreatedTopic = "goal_created"
	GoalDeletedTopic = "goal_deleted"
)

type GoalProducer struct {
	base *messaging.Producer
}

func NewGoalProducer(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *GoalProducer {
	return &GoalProducer{
		base: messaging.NewProducer(producer, logger),
	}
}

func (p *GoalProducer) PublishGoalCreated(
	ctx context.Context,
	event events.GoalEvent,
) error {
	return p.base.Publish(ctx,
		GoalCreatedTopic,
		event,
	)
}

func (p *GoalProducer) PublishGoalDeleted(
	ctx context.Context,
	event events.GoalEvent,
) error {
	return p.base.Publish(ctx,
		GoalDeletedTopic,
		event,
	)
}
