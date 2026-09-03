package transactions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type TransactionRouter struct {
	handler transactionHandler
	m       authMiddleware
}

type transactionHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	FindById(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	DeleteById(w http.ResponseWriter, r *http.Request)
	DeleteByCategoryId(w http.ResponseWriter, r *http.Request)
}

func NewRouter(
	handler transactionHandler,
	m authMiddleware,
) *TransactionRouter {
	return &TransactionRouter{
		handler: handler,
		m:       m,
	}
}

func (r *TransactionRouter) Routes(router chi.Router) {
	router.Route("/transactions", func(router chi.Router) {
		router.Group(func(router chi.Router) {
			router.Use(r.m.RequireActivatedUser)

			router.Post("/", r.handler.Create)
			router.Get("/", r.handler.FindAll)
			router.Get("/{id}", r.handler.FindById)
			router.Put("/{id}", r.handler.Update)
			router.Delete("/{id}", r.handler.DeleteById)
			router.Delete("/category/{id}", r.handler.DeleteByCategoryId)
		})
	})
}
