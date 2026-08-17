package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthRouter struct {
	handler authHandler
}

type authHandler interface {
	Authenticate(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

func NewRouter(handler authHandler) *AuthRouter {
	return &AuthRouter{
		handler: handler,
	}
}

func (r *AuthRouter) Routes(router chi.Router) {
	router.Route("/auth", func(router chi.Router) {
		router.Post("/", r.handler.Authenticate)
		router.Post("/refresh", r.handler.RefreshToken)
		router.Post("/logout", r.handler.Logout)
	})
}
