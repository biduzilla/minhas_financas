package api

import (
	"database/sql"
	"expvar"
	"log/slog"
	"ms_goal/internal/core/config"
	"ms_goal/internal/core/database"
	"ms_goal/internal/core/loggerutils"
	"ms_goal/internal/core/messaging"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type application struct {
	config        config.Config
	Logger        *slog.Logger
	wg            sync.WaitGroup
	db            *sql.DB
	kafkaProducer sarama.SyncProducer
	kafkaConsumer sarama.ConsumerGroup
}

const version = "1.0.0"

func NewApp(cfg config.Config) (*application, error) {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(&loggerutils.ErrorAwareHandler{Handler: baseHandler})

	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info("database connection pool established")

	producers, consumers, err := messaging.InitKafka(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
	)
	if err != nil {
		logger.Error(err.Error(), err)
		return nil, err
	}

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
		config:        cfg,
		Logger:        logger,
		db:            db,
		kafkaProducer: producers,
		kafkaConsumer: consumers,
	}, nil
}
