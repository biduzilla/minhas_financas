package messaging

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Producer struct {
	producer sarama.SyncProducer
	logger   *slog.Logger
	tracer   trace.Tracer
}

func NewProducer(producer sarama.SyncProducer, logger *slog.Logger) *Producer {
	return &Producer{
		producer: producer,
		logger:   logger,
		tracer:   otel.Tracer("messaging/producer"),
	}
}

func (p *Producer) Publish(
	ctx context.Context,
	topic string,
	event Event,
) error {
	spanName := fmt.Sprintf("Publish %s", topic)
	ctx, span := p.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(event.Key()),
		Value:   sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{},
	}

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("failed to publish event",
			"topic", topic,
			"key", event.Key(),
			"error", err,
		)
		return fmt.Errorf("failed to publish event to %s: %w", topic, err)
	}

	p.logger.Info("event published",
		"topic", topic,
		"key", event.Key(),
		"partition", partition,
		"offset", offset,
	)

	return nil
}
