package api

import (
	"database/sql"
	"expvar"
	"fmt"
	"ms_transaction/internal/core/domain/apiError"
	"ms_transaction/internal/core/metrics"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type mw interface {
	Metrics(next http.Handler) http.Handler
	EnableCORS(next http.Handler) http.Handler
	RequireAuthenticatedUser(next http.Handler) http.Handler
	RequireActivatedUser(next http.Handler) http.Handler
	Authenticate(next http.Handler) http.Handler
	RateLimit(next http.Handler) http.Handler
	RecoverPanic(next http.Handler) http.Handler
	Logging(next http.Handler) http.Handler
	TimeoutMiddleWare(next http.Handler) http.Handler
	RequestID(next http.Handler) http.Handler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type Router struct {
	errHandler errorHandler
	m          mw
}

func NewRouter(
	handlers *handlers,
	errHandler errorHandler,
	m mw,
) *Router {
	return &Router{
		m:          m,
		errHandler: errHandler,
	}
}

func (router *Router) RegisterRoutes(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(router.m.RecoverPanic)
	r.Use(router.m.TimeoutMiddleWare)
	r.Use(router.m.RequestID)
	r.Use(router.m.Metrics)
	r.Use(router.m.Logging)

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		router.errHandler.HandlerError(w, req, apiError.ErrRecordNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		message := fmt.Sprintf("the %s method is not supported for this resource", req.Method)
		router.errHandler.HandlerError(w, req, apiError.NewHTTPError(message, http.StatusMethodNotAllowed, nil))
	})

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Handle("/metrics", metrics.MetricsHandler(db))
	r.Mount("/debug/vars", expvar.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(router.m.RateLimit)
		r.Use(router.m.EnableCORS)
		r.Use(router.m.Authenticate)

	})

	return r
}
