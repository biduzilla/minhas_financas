package middleware

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"ms_transaction/internal/core/config"
	"ms_transaction/internal/core/contexts"
	"ms_transaction/internal/core/domain"
	"ms_transaction/internal/core/domain/apiError"
	"ms_transaction/internal/core/metrics"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

var (
	totalRequestsReceived           = expvar.NewInt("total_requests_received")
	totalResponsesSent              = expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
	totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
)

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type JWTService interface {
	ExtractAuthenticatedUser(tokenString string) (domain.UserDetails, error)
}

type Middleware struct {
	errHandler errorHandler
	jwtService JWTService
	config     config.Config
	logger     *slog.Logger
	shutdown   <-chan struct{}
}

func New(
	errHandler errorHandler,
	config config.Config,
	jwtService JWTService,
	logger *slog.Logger,
	shutdown <-chan struct{},
) *Middleware {
	return &Middleware{
		jwtService: jwtService,
		errHandler: errHandler,
		config:     config,
		logger:     logger,
		shutdown:   shutdown,
	}
}

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)
	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (m *Middleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)

		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		d := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(d)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(mw.statusCode)

		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}

func (m *Middleware) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range m.config.CORS.TrustedOrigins {
				if origin == m.config.CORS.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contexts.GetUser(r.Context())

		if user.IsAnonymous() {
			m.errHandler.HandlerError(w, r, apiError.ErrAuthenticationRequired)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequireActivatedUser(next http.Handler) http.Handler {
	return m.RequireAuthenticatedUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contexts.GetUser(r.Context())
		if !user.GetIsAtivo() {
			m.errHandler.HandlerError(w, r, apiError.ErrInactiveAccount)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = r.WithContext(contexts.SetUser(r.Context(), domain.AnonymousUser))
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			m.errHandler.HandlerError(w, r, apiError.ErrInvalidCredentials)
			return
		}

		token := headerParts[1]
		user, err := m.jwtService.ExtractAuthenticatedUser(token)

		if err != nil {
			m.handleTokenError(w, r, err)
			return
		}

		if user == nil {
			m.errHandler.HandlerError(w, r, apiError.ErrInvalidAuthenticationToken)
			return
		}

		if !user.GetIsAtivo() {
			m.errHandler.HandlerError(w, r, apiError.ErrInactiveAccount)
			return
		}

		r = r.WithContext(contexts.SetUser(r.Context(), user))

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) handleTokenError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case strings.Contains(err.Error(), "malformed"):
		m.errHandler.HandlerError(w, r, apiError.ErrMalFormedToken)
	case strings.Contains(err.Error(), "expired"):
		m.errHandler.HandlerError(w, r, apiError.ErrTokenExpired)
	default:
		m.errHandler.HandlerError(w, r, err)
	}
}

func (m *Middleware) RequirePermission(codes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user := contexts.GetUser(r.Context())

			permissions := user.GetRoles()

			hasPermission := false

			for _, code := range codes {
				if slices.Contains(permissions, code) {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.errHandler.HandlerError(w, r, apiError.ErrNotFoundPermission)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for ip, c := range clients {
					if time.Since(c.lastSeen) > 3*time.Minute {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			case <-m.shutdown:
				mu.Lock()
				clients = make(map[string]*client)
				mu.Unlock()
				return
			}
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.config.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				m.errHandler.HandlerError(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(
						rate.Limit(m.config.Limiter.RPS),
						m.config.Limiter.Burst,
					),
				}
			}
			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				m.errHandler.HandlerError(w, r, apiError.ErrRateLimitExceeded)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				m.errHandler.HandlerError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mw := newMetricsResponseWriter(w)

		next.ServeHTTP(mw, r)

		m.logger.Info("request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"status", http.StatusText(mw.statusCode),
			"duration", time.Since(start).String(),
			"request_id", contexts.GetRequestID(r.Context()),
		)
	})
}

func (m *Middleware) TimeoutMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), m.config.Server.Timeout)
		defer cancel()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-Id", id)
		r = r.WithContext(contexts.SetRequestID(r.Context(), id))
		next.ServeHTTP(w, r)
	})
}
