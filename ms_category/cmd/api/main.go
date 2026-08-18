package main

import (
	"log/slog"
	"ms_category/internal/api"
	"ms_category/internal/core/config"
	"os"
	"time"
)

func main() {
	// cfg := config.New()
	cfg := config.Config{}
	cfg.Env = "development"

	cfg.Server.Port = 4002
	cfg.Server.Timeout = 5 * time.Second

	cfg.DB.DSN = "postgres://api_user:api_password@localhost:5432/api_db?sslmode=disable"
	cfg.DB.MaxOpenConns = 25
	cfg.DB.MaxIdleConns = 25
	cfg.DB.MaxIdleTime = "15m"

	cfg.Limiter.RPS = 2.0
	cfg.Limiter.Burst = 4
	cfg.Limiter.Enabled = true

	cfg.Security.PrivateKeyPath = "resources/keys/privateKey.pem"
	cfg.Security.PublicKeyPath = "resources/keys/publicKey.pem"

	cfg.Cache.Addr = "localhost:6379"
	cfg.Cache.Password = "redis_secure_password"
	cfg.Cache.Db = 0

	cfg.Otel.Port = "localhost:4318"

	app := api.NewApp(cfg)
	if app == nil {
		slog.Error("failed to initialize app")
		os.Exit(1)
	}
	err := app.Server()
	if err != nil {
		app.Logger.Error(err.Error())
	}
}
