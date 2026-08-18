package api

import (
	"log/slog"
	"ms_category/internal/core/cache"
	"ms_category/internal/core/config"
	"ms_category/internal/core/security"
	"ms_category/internal/core/transaction"
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
	cacheClient, err := cache.NewRedisCache(config.Cache.Addr, config.Cache.Password, config.Cache.Db, nil)

	if err != nil {
		return nil, err
	}

	logger.Info("reddis connection pool established")

	return &services{}, nil
}
