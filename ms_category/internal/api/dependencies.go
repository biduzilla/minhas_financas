package api

import (
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/core/middleware"
	"ms_category/internal/core/transaction"
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

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, func(), error) {
	repo := NewRepositories(app.db, app.Logger)
	tx := transaction.NewManager(app.db)
	clients := NewClients(app.config)
	services, err := NewServices(repo, tx, app.config, app.Logger, clients)
	if err != nil {
		return nil, nil, err
	}

	consumers := NewConsumer(app.kafkaConsumer, services, app.Logger)
	shutdownConsumers := consumers.Start(app.Logger)
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
	}, shutdownConsumers, nil
}
