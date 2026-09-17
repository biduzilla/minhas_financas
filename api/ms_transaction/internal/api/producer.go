package api

import (
	"log/slog"
	p "ms_transaction/internal/core/messaging/producers"

	"github.com/IBM/sarama"
)

type producers struct {
	transactionProducer *p.TransactionProducer
}

func NewProducers(
	producer sarama.SyncProducer,
	logger *slog.Logger,
) *producers {
	return &producers{
		transactionProducer: p.NewTransactionProducer(producer, logger),
	}
}
