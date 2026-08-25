package api

import (
	"log/slog"
	"ms_goal/internal/core/config"
	"ms_goal/internal/core/security"
	"ms_goal/internal/core/transaction"
)

type services struct {
	jwtService *security.JwtService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
) (*services, error) {
	// cacheClient, err := cache.NewRedisCache(config.Cache.Addr, config.Cache.Password, config.Cache.Db, nil)

	// if err != nil {
	// 	return nil, err
	// }

	logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(config)
	if err != nil {
		return nil, err
	}

	// transactionKeyBuilder := cache.NewKeyBuilder("transactions")
	// transactionWriteExecutor := transaction.NewWriterExecutor(tx, cacheClient, transactionKeyBuilder)
	// transactionService := transactions.NewService(
	// 	r.transaction,
	// 	cacheClient,
	// 	transactionKeyBuilder,
	// 	transactionWriteExecutor,
	// )

	return &services{
		jwtService: jwtService,
	}, nil
}
