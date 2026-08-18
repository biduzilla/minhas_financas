package api

import (
	"database/sql"
	"expvar"
	"log/slog"
	"ms_category/internal/core/config"
	"ms_category/internal/core/database"
	"ms_category/internal/core/loggerutils"
	"os"
	"runtime"
	"sync"
	"time"
)

type application struct {
	config config.Config
	Logger *slog.Logger
	wg     sync.WaitGroup
	db     *sql.DB
}

const version = "1.0.0"

func NewApp(cfg config.Config) *application {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(&loggerutils.ErrorAwareHandler{Handler: baseHandler})

	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	if err := database.RunMigrations(cfg.DB.DSN, logger); err != nil {
		logger.Error("failed to run migrations", "error", err)
		return nil
	}

	logger.Info("database connection pool established")

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	return &application{
		config: cfg,
		Logger: logger,
		db:     db,
	}
}
