package auth

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/handler"
	"ms_auth/pkg/httputil"
)

type AuthHandler struct {
	authService authService
	errHandler  errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type authService interface {
	Authenticate(ctx context.Context, email, password string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

func NewHandler(
	authService authService,
	errHandler errorHandler,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		errHandler:  errHandler,
	}
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "AuthHandler.Authenticate")
	defer span.End()

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httputil.ReadJSON(w, r, &input); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apiError.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	token, err := h.authService.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Authentication failed")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Authentication successful")
	handler.Respond(w, r, http.StatusOK, token, nil, h.errHandler)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "AuthHandler.RefreshToken")
	defer span.End()

	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := httputil.ReadJSON(w, r, &input); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apiError.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	token, err := h.authService.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Refresh token failed")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Token refreshed")
	handler.Respond(w, r, http.StatusOK, token, nil, h.errHandler)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "AuthHandler.Logout")
	defer span.End()

	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := httputil.ReadJSON(w, r, &input); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apiError.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	if err := h.authService.Logout(ctx, input.RefreshToken); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Logout failed")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Logout successful")
	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
