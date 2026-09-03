package api

import (
	"log/slog"
	"ms_goal/internal/core/cache"
	"ms_goal/internal/core/config"
	"ms_goal/internal/core/security"
	"ms_goal/internal/core/transaction"
	"ms_goal/internal/features/goal"
	goaltransaction "ms_goal/internal/features/goal_transaction"
)

type services struct {
	jwtService      *security.JwtService
	goal            *goal.GoalService
	goalTransaction *goaltransaction.GoalTransactionService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
	producers *producers,
) (*services, error) {
	cacheClient, err := cache.NewRedisCache(config.Cache.Addr, config.Cache.Password, config.Cache.Db, nil)

	if err != nil {
		return nil, err
	}

	logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(config)
	if err != nil {
		return nil, err
	}

	gTd := newWriteDeps("goals_transactions", tx, cacheClient)
	goalTService := goaltransaction.NewService(
		r.goalTransaction,
		gTd.cache,
		gTd.keyBuilder,
		gTd.executor,
	)

	gD := newWriteDeps("goals", tx, cacheClient)
	goalService := goal.NewService(
		r.goal,
		gD.cache,
		gD.keyBuilder,
		gD.executor,
		producers.goalProducer,
		goalTService,
	)

	return &services{
		jwtService:      jwtService,
		goal:            goalService,
		goalTransaction: goalTService,
	}, nil
}

type writeDeps struct {
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	executor   *transaction.WriteExecutor
}

func newWriteDeps(prefix string, tx transaction.Manager, cacheClient cache.Cache) writeDeps {
	kb := cache.NewKeyBuilder(prefix)
	we := transaction.NewWriterExecutor(tx, cacheClient, kb)
	return writeDeps{
		cache:      cacheClient,
		keyBuilder: kb,
		executor:   we,
	}
}
