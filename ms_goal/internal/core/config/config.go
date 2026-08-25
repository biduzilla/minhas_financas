package config

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	Server struct {
		Port    int           `env:"SERVER_PORT,required"`
		Timeout time.Duration `env:"SERVER_TIMEOUT,required"`
	}
	Env string
	DB  struct {
		DSN          string `env:"DB_DSN,required"`
		MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS,required"`
		MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS,required"`
		MaxIdleTime  string `env:"DB_MAX_IDLE_TIME,required"`
	}
	Limiter struct {
		RPS     float64 `env:"LIMITER_RPS,required"`
		Burst   int     `env:"LIMITER_BURST,required"`
		Enabled bool    `env:"LIMITER_ENABLED,required"`
	}
	CORS struct {
		TrustedOrigins []string
	}
	Security struct {
		PrivateKeyPath string `env:"PRIVATE_KEY_PATH,required"`
		PublicKeyPath  string `env:"PUBLIC_KEY_PATH,required"`
	}
	Cache struct {
		Addr     string `env:"CACHE_ADDR,required"`
		Password string `env:"CACHE_PASSWORD,required"`
		Db       int    `env:"CACHE_DB,required"`
	}
	Otel struct {
		Port string `env:"JAEGER_PORT,required"`
	}
}

func New() *Config {
	var c Config
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
