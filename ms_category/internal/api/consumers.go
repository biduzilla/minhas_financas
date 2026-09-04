package api

import (
	"context"
	"log/slog"
	"ms_category/internal/core/messaging"
	"ms_category/internal/core/messaging/events"
	c "ms_category/internal/features/consumers"

	"github.com/IBM/sarama"
)

type consumers struct {
	kafkaConsumer *messaging.Consumer
}

func NewConsumer(
	consumerGroup sarama.ConsumerGroup,
	services *services,
	logger *slog.Logger,
) *consumers {
	goalHandler := c.NewGoalEventHandler(
		services.category,
		logger,
	)

	handlersMap := map[string]messaging.EventHandler{
		events.GoalCreatedTopic: goalHandler,
		events.GoalDeletedTopic: goalHandler,
	}

	return &consumers{
		kafkaConsumer: messaging.NewConsumer(consumerGroup, handlersMap, logger),
	}
}

func (c *consumers) Start(logger *slog.Logger) (shutdownFunc func()) {
	logger.Info("starting kafka consumers...")

	consumerCtx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		if err := c.kafkaConsumer.Start(consumerCtx); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	shutdownFunc = func() {
		logger.Info("stopping kafka consumers...")
		cancel()
		<-errCh
		logger.Info("kafka consumers stopped")
	}

	go func() {
		if err := <-errCh; err != nil {
			logger.Error("kafka consumer crashed", "error", err)
		}
	}()

	return shutdownFunc
}
