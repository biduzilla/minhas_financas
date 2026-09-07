package config

import (
	"log"
	"shared/config"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Base  config.Config
	Kafka struct {
		Brokers []string `env:"KAFKA_BROKERS,required"`
		GroupID string   `env:"KAFKA_GROUP_ID,required"`
	}
	Clients struct {
		CategoryURL string `env:"CATEGORY_SERVICE_URL,required"`
	}
}

func New() *Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
