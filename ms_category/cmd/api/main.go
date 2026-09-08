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
	cfg.Base.Env = "development"

	cfg.Base.Server.Port = 4002
	cfg.Base.Server.Timeout = 5 * time.Second

	cfg.Base.DB.DSN = "postgres://api_user:api_password@localhost:5432/api_db?sslmode=disable"
	cfg.Base.DB.MaxOpenConns = 25
	cfg.Base.DB.MaxIdleConns = 25
	cfg.Base.DB.MaxIdleTime = "15m"

	cfg.Base.Limiter.RPS = 2.0
	cfg.Base.Limiter.Burst = 4
	cfg.Base.Limiter.Enabled = true

	cfg.Base.Security.PrivateKeyPath = "resources/keys/privateKey.pem"
	cfg.Base.Security.PublicKeyPath = "resources/keys/publicKey.pem"

	cfg.Base.Cache.Addr = "localhost:6379"
	cfg.Base.Cache.Password = "redis_secure_password"
	cfg.Base.Cache.Db = 0

	cfg.Base.Otel.Port = "localhost:4318"

	app, err := api.NewApp(cfg)
	if err != nil {
		slog.Error("failed to initialize app")
		os.Exit(1)
	}

	err = app.Server()
	if err != nil {
		app.Logger.Error(err.Error())
	}
}
