package api

import (
	"log/slog"
	"ms_transaction/internal/core/config"
	"ms_transaction/internal/core/security"
	"ms_transaction/internal/core/transaction"
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

	return &services{
		jwtService: jwtService,
	}, nil
}
