package category

import (
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/core/filters"
	"ms_category/internal/core/handler"
	"ms_category/internal/core/validator"
	"ms_category/pkg/httpjson"
	"ms_category/pkg/httputil"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type CategoryHandler struct {
	service    service
	errHandler errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func NewHandler(
	service service,
	errHandler errorHandler,
) *CategoryHandler {
	return &CategoryHandler{
		service:    service,
		errHandler: errHandler,
	}
}

type categoryHandler interface {
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

func (h *CategoryHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_category/internal/features/category")
	ctx, span := tracer.Start(r.Context(), "CategoryHandler.Create")

	defer span.End()

	var dto CategoryDTO
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

	span.SetAttributes(attribute.String("category.Name", *dto.Name))

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

func (h *CategoryHandler) FindById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_category/internal/features/category")
	ctx, span := tracer.Start(r.Context(), "CategoryHandler.FindById")
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

func (h *CategoryHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_category/internal/features/category")
	ctx, span := tracer.Start(r.Context(), "CategoryHandler.FindAll")

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

	models, metadata, err := h.service.FindAll(ctx,
		input.search,
		input.Filters,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch users")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	dtos := make([]CategoryDTO, len(models))

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

func (h *CategoryHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_category/internal/features/category")
	ctx, span := tracer.Start(r.Context(), "CategoryHandler.Update")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	var dto CategoryDTO
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

	span.SetAttributes(attribute.String("category.id", id.String()))

	if err := h.service.Update(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update user")
		h.errHandler.HandlerError(w, r, err)
		return
	}
	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *CategoryHandler) DeleteById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_category/internal/features/category")
	ctx, span := tracer.Start(r.Context(), "CategoryHandler.DeleteById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("category.id", id.String()))

	if err := h.service.DeleteById(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete user")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
