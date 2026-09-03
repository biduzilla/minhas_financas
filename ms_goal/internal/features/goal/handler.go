package goal

import (
	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/core/filters"
	"ms_goal/internal/core/handler"
	"ms_goal/internal/core/validator"
	"ms_goal/pkg/httpjson"
	"ms_goal/pkg/httputil"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GoalHandler struct {
	service    service
	errHandler errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func NewHandler(
	service service,
	errHandler errorHandler,
) *GoalHandler {
	return &GoalHandler{
		service:    service,
		errHandler: errHandler,
	}
}

type goalHandler interface {
	Create(
		w http.ResponseWriter,
		r *http.Request,
	)
	FindById(
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

func (h *GoalHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_goal/internal/features/goal")
	ctx, span := tracer.Start(r.Context(), "GoalHandler.Create")

	defer span.End()

	var dto GoalDTO
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

	span.SetAttributes(attribute.String("goal.Name", *dto.Name))

	model := dto.ToModel()

	err := h.service.Insert(ctx, model)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusCreated, model.ToDTO(), nil, h.errHandler)
}

func (h *GoalHandler) FindById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_goal/internal/features/goal")
	ctx, span := tracer.Start(r.Context(), "GoalHandler.FindById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	model, err := h.service.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch by ID")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *GoalHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_goal/internal/features/goal")
	ctx, span := tracer.Start(r.Context(), "GoalHandler.FindAll")

	defer span.End()

	var input struct {
		search string
		status GoalStatus
		filters.Filters
	}

	v := validator.New()

	s := httputil.ReadStringParam(r, "status", "")
	status, err := ParseGoalStatus(s)
	if err != nil {
		h.errHandler.HandlerError(
			w,
			r,
			apiError.NewBadRequestError(err))
		return
	}
	input.status = status
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

	models, metadata, err := h.service.FindAll(ctx,
		input.search,
		input.status,
		input.Filters,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch users")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	dtos := make([]GoalDTO, len(models))

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

func (h *GoalHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_goal/internal/features/goal")
	ctx, span := tracer.Start(r.Context(), "GoalHandler.Update")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	var dto GoalDTO
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

	span.SetAttributes(attribute.String("goal.id", id.String()))

	if err := h.service.Update(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update user")
		h.errHandler.HandlerError(w, r, err)
		return
	}
	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *GoalHandler) DeleteById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_goal/internal/features/goal")
	ctx, span := tracer.Start(r.Context(), "GoalHandler.DeleteById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("goal.id", id.String()))

	if err := h.service.DeleteById(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete user")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
