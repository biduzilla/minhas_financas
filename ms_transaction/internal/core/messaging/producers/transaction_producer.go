package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"ms_transaction/internal/core/messaging/events"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	TransactionGoalCreatedTopic = "transaction_goal_created"
)

type TransactionProducer struct {
	producer sarama.SyncProducer
	logger   *slog.Logger
}

func NewTransactionProducer(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *TransactionProducer {
	return &TransactionProducer{
		producer: producer,
		logger:   logger,
	}
}

func (p *TransactionProducer) PublishTransactionGoalCreated(
	ctx context.Context,
	event events.TransactionEvent,
) error {
	return p.publishEvent(ctx,
		event,
		TransactionGoalCreatedTopic,
		"TransactionProducer.PublishTransactionGoalCreated",
	)
}

func (p *TransactionProducer) publishEvent(
	ctx context.Context,
	event events.TransactionEvent,
	topic string,
	spanName string,
) error {
	tracer := otel.Tracer("ms_transaction/internal/core/messaging/producers")
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
			"Failed to "+topic,
			"eventId", event.ID.String(),
		)

		return fmt.Errorf("failed to %s: %w", topic, err)
	}

	p.logger.Info(topic+" published",
		"eventId", event.ID.String(),
		"partition", string(partition),
		"offset", fmt.Sprint(offset),
	)

	return nil
}
