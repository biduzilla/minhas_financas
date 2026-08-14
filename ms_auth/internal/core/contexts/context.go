package contexts

import (
	"context"
	"database/sql"
	"ms_auth/internal/core/domain"
)

type contextKey string

const userContextKey = contextKey("user")
const txContextKey = contextKey("tx")
const requestIDKey = contextKey("request_id")

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func SetUser(ctx context.Context, user domain.UserDetails) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) domain.UserDetails {
	user, ok := ctx.Value(userContextKey).(domain.UserDetails)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

func SetTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}

func GetTx(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txContextKey).(*sql.Tx)
	return tx
}
