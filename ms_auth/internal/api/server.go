package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"ms_auth/internal/core/otel"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type router interface {
	RegisterRoutes(db *sql.DB) *chi.Mux
}

func (app *application) Server() error {
	defer app.db.Close()

	shutdown := make(chan struct{})

	shutdownTracer, err := otel.InitTracer("ms_auth", app.Logger)
	if err != nil {
		return err
	}

	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			app.Logger.Error("failed to shutdown tracer", "error", err)
		}
	}()

	deps, err := app.buildDependencies(shutdown)
	if err != nil {
		return err
	}

	mux := deps.routers.RegisterRoutes(app.db)
	instrumentedHandler := otelhttp.NewHandler(mux, "ms_auth")

	srv := newHTTPServer(
		fmt.Sprintf(":%d", app.config.Server.Port),
		instrumentedHandler,
		app.Logger,
	)

	shutdownError := make(chan error, 1)
	app.handleShutdown(srv, shutdown, shutdownError)

	app.Logger.Info("starting server", "addr", srv.Addr)

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.Info("stopped server", "addr", srv.Addr)
	return nil
}

func newHTTPServer(addr string, handler http.Handler, logger *slog.Logger) *http.Server {
	errorLog := slog.NewLogLogger(logger.Handler(), slog.LevelError)
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ErrorLog:     errorLog,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func (app *application) handleShutdown(srv *http.Server, shutdown chan struct{}, shutdownError chan<- error) {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.Info("shutting down", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		close(shutdown)

		app.Logger.Info("completing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()
}
