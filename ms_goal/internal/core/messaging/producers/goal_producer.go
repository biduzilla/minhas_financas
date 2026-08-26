package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"ms_goal/internal/core/messaging/events"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	GoalCreatedTopic = "goal_created"
	GoalDeletedTopic = "goal_deleted"
)

type GoalProducer struct {
	producer sarama.SyncProducer
	logger   *slog.Logger
}

func NewGoalProducer(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *GoalProducer {
	return &GoalProducer{
		producer: producer,
		logger:   logger,
	}
}

func (p *GoalProducer) PublishGoalCreated(
	ctx context.Context,
	event events.GoalEvent,
) error {
	return p.publishEvent(ctx,
		event,
		GoalCreatedTopic,
		"GoalProducer.PublishGoalCreated",
	)
}

func (p *GoalProducer) PublishGoalDeleted(
	ctx context.Context,
	event events.GoalEvent,
) error {
	return p.publishEvent(ctx,
		event,
		GoalDeletedTopic,
		"GoalProducer.PublishGoalDeleted",
	)
}

func (p *GoalProducer) publishEvent(
	ctx context.Context,
	event events.GoalEvent,
	topic string,
	spanName string,
) error {
	tracer := otel.Tracer("ms_goal/internal/core/messaging/producers")
	ctx, span := tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal goal created event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(event.ID.String()),
		Value:   sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{},
	}

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		msg.Headers = append(
			msg.Headers,
			sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			},
		)
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error(
			"Failed to publish order created event",
			"eventId", event.ID.String(),
		)

		return fmt.Errorf("failed to publish goal created event: %w", err)
	}

	p.logger.Info("Goal created event published",
		"eventId", event.ID.String(),
		"partition", string(partition),
		"offset", fmt.Sprint(offset),
	)

	return nil
}
