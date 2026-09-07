package producers

import (
	"context"
	"log/slog"
	"messaging"
	"ms_transaction/internal/core/messaging/events"

	"github.com/IBM/sarama"
)

const (
	TransactionGoalCreatedTopic = "transaction_goal_created"
	TransactionGoalDeletedTopic = "transaction_goal_deleted"
)

type TransactionProducer struct {
	base *messaging.Producer
}

func NewTransactionProducer(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *TransactionProducer {
	return &TransactionProducer{
		base: messaging.NewProducer(producer, logger),
	}
}

func (p *TransactionProducer) PublishTransactionGoalCreated(
	ctx context.Context,
	event events.TransactionEvent,
) error {
	return p.base.Publish(ctx,
		TransactionGoalCreatedTopic,
		event,
	)
}

func (p *TransactionProducer) PublishTransactionGoalDeleted(
	ctx context.Context,
	event events.TransactionEvent,
) error {
	return p.base.Publish(ctx,
		TransactionGoalDeletedTopic,
		event,
	)
}
