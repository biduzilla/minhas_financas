package messaging

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

func InitKafka(brokers []string, groupID string) (sarama.SyncProducer, sarama.ConsumerGroup, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Net.MaxOpenRequests = 1
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Idempotent = true

	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false

	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}

	saramaCfg.Net.DialTimeout = 3 * time.Second
	saramaCfg.Net.ReadTimeout = 5 * time.Second
	saramaCfg.Net.WriteTimeout = 5 * time.Second

	var producer sarama.SyncProducer
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		producer, err = sarama.NewSyncProducer(brokers, saramaCfg)
		if err == nil {
			break
		}
		fmt.Printf("Aguardando Kafka iniciar... tentativa %d/%d (erro: %v)\n", i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to start kafka producer after retries: %w", err)
	}

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, saramaCfg)
	if err != nil {
		producer.Close()
		return nil, nil, fmt.Errorf("failed to start kafka consumer: %w", err)
	}

	return producer, consumer, nil
}
