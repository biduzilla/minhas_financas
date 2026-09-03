package api

import (
	"log/slog"
	p "ms_goal/internal/core/messaging/producers"

	"github.com/IBM/sarama"
)

type producers struct {
	goalProducer *p.GoalProducer
}

func NewProducers(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *producers {
	return &producers{
		goalProducer: p.NewGoalProducer(producer, logger),
	}
}
