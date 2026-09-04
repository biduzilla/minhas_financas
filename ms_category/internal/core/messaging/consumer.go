package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type EventHandler interface {
	Handle(ctx context.Context, message []byte, topic string) error
}

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topicHandlers map[string]EventHandler
	tracer        trace.Tracer
	logger        *slog.Logger
}

func NewConsumer(
	consumerGroup sarama.ConsumerGroup,
	topicHandlers map[string]EventHandler,
	logger *slog.Logger,
) *Consumer {
	return &Consumer{
		consumerGroup: consumerGroup,
		topicHandlers: topicHandlers,
		tracer:        otel.Tracer("ms_category/internal/core/messaging"),
		logger:        logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := make([]string, 0, len(c.topicHandlers))
	for topic := range c.topicHandlers {
		topics = append(topics, topic)
	}

	for {
		err := c.consumerGroup.Consume(ctx, topics, c)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			c.logger.Warn("kafka consumer error, retrying...", "error", err.Error())

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	c.logger.Info("✅ Kafka consumer connected successfully")
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	c.logger.Info("Kafka consumer disconnected")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if len(c.topicHandlers) == 0 {
		return fmt.Errorf("no topic handlers registered")
	}

	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			carrier := propagation.MapCarrier{}
			for _, h := range msg.Headers {
				carrier[string(h.Key)] = string(h.Value)
			}

			ctx := otel.GetTextMapPropagator().Extract(session.Context(), carrier)

			ctx, span := c.tracer.Start(ctx, "Kafka.ConsumeMessage",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					attribute.String("messaging.system", "kafka"),
					attribute.String("messaging.destination", msg.Topic),
					attribute.String("messaging.message_id", string(msg.Key)),
				),
			)

			handler, exists := c.topicHandlers[msg.Topic]
			if !exists {
				c.logger.Error("No handler registered for topic", "topic", msg.Topic)
				span.SetStatus(codes.Error, "no handler found")
				session.MarkMessage(msg, "")
				span.End()
				continue
			}

			err := handler.Handle(ctx, msg.Value, msg.Topic)

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				c.logger.Error("Failed to process message",
					"topic", msg.Topic,
					"error", err,
				)
				span.End()
				return err
			}

			span.SetStatus(codes.Ok, "processed")
			session.MarkMessage(msg, "")
			span.End()
		}
	}
}
