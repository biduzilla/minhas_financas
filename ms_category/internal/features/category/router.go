package category

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type CategoryRouter struct {
	handler categoryHandler
	m       authMiddleware
}

func NewRouter(
	handler categoryHandler,
	m authMiddleware,
) *CategoryRouter {
	return &CategoryRouter{
		handler: handler,
		m:       m,
	}
}

func (r *CategoryRouter) Routes(router chi.Router) {
	router.Route("/categories", func(router chi.Router) {

		router.Group(func(router chi.Router) {
			router.Use(r.m.RequireActivatedUser)

			router.Post("/", r.handler.Create)
			router.Get("/{id}", r.handler.FindById)
			router.Get("/", r.handler.FindAll)
			router.Put("/{id}", r.handler.Update)
			router.Delete("/{id}", r.handler.DeleteById)
		})
	})
}
