package config

import (
	"log"
	"shared/config"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Base    config.Config
	Clients struct {
		TransactionURL string `env:"TRANSACTION_SERVICE_URL,required"`
	}
	Kafka struct {
		Brokers []string `env:"KAFKA_BROKERS,required"`
		GroupID string   `env:"KAFKA_GROUP_ID,required"`
	}
}

func New() *Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
