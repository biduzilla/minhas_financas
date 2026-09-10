package api

import (
	"log/slog"
	"ms_goal/internal/core/config"
	"ms_goal/internal/features/goal"
	goaltransaction "ms_goal/internal/features/goal_transaction"
	"shared/auth/security"
	"shared/cache"
	"shared/transaction"
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
	cacheClient, err := cache.NewRedisCache(
		config.Base.Cache.Addr,
		config.Base.Cache.Password,
		config.Base.Cache.Db,
		nil,
	)

	if err != nil {
		return nil, err
	}

	logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	gTd := newWriteDeps("goals_transactions", tx, cacheClient)
	goalTService := goaltransaction.NewService(
		r.goalTransaction,
		gTd.cache,
		gTd.keyBuilder,
		gTd.executor,
		nil,
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

	goalTService.SetGoalService(goalService)

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
