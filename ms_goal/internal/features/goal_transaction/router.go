package goaltransaction

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type GoalTransactionRouter struct {
	handler goalTransactionHandler
	m       authMiddleware
}

func NewRouter(
	handler goalTransactionHandler,
	m authMiddleware,
) *GoalTransactionRouter {
	return &GoalTransactionRouter{
		handler: handler,
		m:       m,
	}
}

func (r *GoalTransactionRouter) Routes(router chi.Router) {
	router.Route("/goals_transactions", func(router chi.Router) {

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
