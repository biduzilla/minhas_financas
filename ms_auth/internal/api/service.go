package api

import (
	"log/slog"
	"ms_auth/internal/core/cache"
	"ms_auth/internal/core/config"
	"ms_auth/internal/core/security"
	"ms_auth/internal/core/transaction"
	"ms_auth/internal/features/auth"
	"ms_auth/internal/features/user"
)

type services struct {
	jwtService *security.JwtService
	user       *user.UserService
	auth       *auth.AuthService
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

	jwtService, err := security.NewService(config)
	if err != nil {
		return nil, err
	}

	userKeyBuilder := cache.NewKeyBuilder("user")
	userWriteExecutor := transaction.NewWriterExecutor(tx, cacheClient, userKeyBuilder)
	userService := user.NewService(
		r.user,
		cacheClient,
		userKeyBuilder,
		userWriteExecutor,
	)

	return &services{
		jwtService: jwtService,
		user:       userService,
		auth:       auth.NewService(jwtService, userService, r.auth, tx),
	}, nil
}
