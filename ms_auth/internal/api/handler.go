package api

import (
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/features/user"
)

type handlers struct {
	user *user.UserHandler
}

func NewHandlers(
	services *services,
	errHandler *apiError.ErrorHandler,
) *handlers {
	return &handlers{
		user: user.NewHandler(services.user, errHandler),
	}
}
