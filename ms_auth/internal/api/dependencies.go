package api

import (
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/middleware"
	"ms_auth/internal/core/transaction"
)

type dependencies struct {
	repo         *repositories
	tx           transaction.Manager
	services     *services
	errorHandler *apiError.ErrorHandler
	handlers     *handlers
	mw           *middleware.Middleware
	routers      *Router
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
	repo := NewRepositories(app.db, app.Logger)
	tx := transaction.NewManager(app.db)
	services, err := NewServices(repo, tx, app.config, app.Logger)
	if err != nil {
		return nil, err
	}

	errHandler := apiError.NewErrorHandler(app.Logger)
	handlers := NewHandlers(services, errHandler)
	middleware := middleware.New(
		errHandler,
		app.config,
		services.jwtService,
		app.Logger,
		shutdown,
	)

	router := NewRouter(
		handlers,
		errHandler,
		middleware,
	)

	return &dependencies{
		repo:         repo,
		tx:           tx,
		services:     services,
		errorHandler: errHandler,
		handlers:     handlers,
		mw:           middleware,
		routers:      router,
	}, nil
}
