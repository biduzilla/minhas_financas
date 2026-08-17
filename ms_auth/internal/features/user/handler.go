package user

import (
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/filters"
	"ms_auth/internal/core/handler"
	"ms_auth/internal/core/validator"
	"ms_auth/pkg/httpjson"
	"ms_auth/pkg/httputil"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type UserHandler struct {
	userService userService
	errHandler  errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type userHandler interface {
	SignUp(
		w http.ResponseWriter,
		r *http.Request,
	)

	FindAll(
		w http.ResponseWriter,
		r *http.Request,
	)

	Update(
		w http.ResponseWriter,
		r *http.Request,
	)

	DeleteById(
		w http.ResponseWriter,
		r *http.Request,
	)
}

func NewHandler(
	userService userService,
	errHandler errorHandler,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		errHandler:  errHandler,
	}
}

func (h *UserHandler) SignUp(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "UserHandler.SignUp")

	defer span.End()

	var dto CreateUserDTO
	if err := httputil.ReadJSON(w, r, &dto); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apiError.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	span.SetAttributes(attribute.String("user.Email", dto.Email))
	span.SetAttributes(attribute.String("user.Name", dto.Name))

	user, err := h.userService.SignUp(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusCreated, user.ToDTO(), nil, h.errHandler)
}

func (h *UserHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "UserHandler.FindAll")

	defer span.End()

	var input struct {
		search string
		filters.Filters
	}

	v := validator.New()

	input.search = httputil.ReadStringParam(r, "search", "")
	input.Filters.Page = httputil.ReadIntParam(r, "page", 1, v)
	input.Filters.PageSize = httputil.ReadIntParam(r, "page_size", 20, v)
	input.Filters.Sort = httputil.ReadStringParam(r, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		h.errHandler.HandlerError(
			w,
			r,
			apiError.NewValidationError(v.Errors))
		return
	}

	models, metadata, err := h.userService.FindAll(ctx,
		input.search,
		input.Filters,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch users")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	dtos := make([]UserDTO, len(models))

	for i, m := range models {
		dtos[i] = m.ToDTO()
	}

	handler.Respond(
		w,
		r,
		http.StatusOK,
		httpjson.Envelope{
			"content":  dtos,
			"metadata": metadata,
		},
		nil,
		h.errHandler,
	)
}

func (h *UserHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "UserHandler.Update")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	var dto UserDTO
	if err := httputil.ReadJSON(w, r, &dto); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apiError.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	model := dto.ToModel()
	model.ID = id

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := h.userService.Update(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update user")
		h.errHandler.HandlerError(w, r, err)
		return
	}
	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *UserHandler) DeleteById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "UserHandler.DeleteById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := h.userService.DeleteById(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete user")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
