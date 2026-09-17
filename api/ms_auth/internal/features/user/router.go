package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type UserRouter struct {
	handler userHandler
	m       authMiddleware
}

func NewRouter(
	handler userHandler,
	m authMiddleware,
) *UserRouter {
	return &UserRouter{
		handler: handler,
		m:       m,
	}
}

func (r *UserRouter) Routes(router chi.Router) {
	router.Route("/users", func(router chi.Router) {
		router.Post("/", r.handler.SignUp)

		router.Group(func(router chi.Router) {
			router.Use(r.m.RequireActivatedUser)

			router.Get("/", r.handler.FindAll)
			router.Put("/{id}", r.handler.Update)
			router.Delete("/{id}", r.handler.DeleteById)
		})
	})
}
