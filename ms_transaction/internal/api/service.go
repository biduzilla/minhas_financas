package api

import (
	"log/slog"
	"ms_transaction/internal/features/transactions"
	"shared/cache"
	"shared/config"
	"shared/security"
	"shared/transaction"
)

type services struct {
	jwtService  *security.JwtService
	transaction *transactions.TransactionService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
	clients *clients,
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

	transactionKeyBuilder := cache.NewKeyBuilder("transactions")
	transactionWriteExecutor := transaction.NewWriterExecutor(tx, cacheClient, transactionKeyBuilder)
	transactionService := transactions.NewService(
		r.transaction,
		cacheClient,
		transactionKeyBuilder,
		transactionWriteExecutor,
		clients.category,
		producers.transactionProducer,
	)

	return &services{
		jwtService:  jwtService,
		transaction: transactionService,
	}, nil
}
