package config

import (
	"log"
	"shared/config"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Base config.Config
}

func New() *Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
