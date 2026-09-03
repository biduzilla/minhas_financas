package api

import (
	"log/slog"
	"ms_category/internal/core/cache"
	"ms_category/internal/core/config"
	"ms_category/internal/core/security"
	"ms_category/internal/core/transaction"
	"ms_category/internal/features/category"
)

type services struct {
	jwtService *security.JwtService
	category   *category.CategoryService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
	clients *clients,
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

	categoryKeyBuilder := cache.NewKeyBuilder("categories")
	categoryWriteExecutor := transaction.NewWriterExecutor(tx, cacheClient, categoryKeyBuilder)
	categoryService := category.NewService(
		r.category,
		cacheClient,
		categoryKeyBuilder,
		categoryWriteExecutor,
		clients.transactions,
	)

	return &services{
		jwtService: jwtService,
		category:   categoryService,
	}, nil
}
