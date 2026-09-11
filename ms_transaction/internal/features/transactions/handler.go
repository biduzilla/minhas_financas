package transactions

import (
	"net/http"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"shared/auth/domain/apiError"
	"shared/httpx/handler"
	"shared/httpx/httputil"
	"shared/utils/filters"
	"shared/validator"
)

type TransactionHandler struct {
	service    service
	errHandler errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func NewHandler(
	service service,
	errHandler errorHandler,
) *TransactionHandler {
	return &TransactionHandler{
		service:    service,
		errHandler: errHandler,
	}
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.Create")
	defer span.End()

	var dto TransactionDTO
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
	span.SetAttributes(
		attribute.String("transaction.category_id", model.CategoryID.String()),
		attribute.Float64("transaction.amount", model.Amount),
	)

	if err := h.service.Insert(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create transaction")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusCreated, model.ToDTO(), nil, h.errHandler)
}

func (h *TransactionHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.FindAll")
	defer span.End()

	query, err := parseTransactionQuery(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid query parameters")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	models, metadata, err := h.service.FindAll(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch transactions")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	dtos := make([]TransactionDTO, len(models))
	for i, m := range models {
		dtos[i] = m.ToDTO()
	}

	handler.Respond(w, r, http.StatusOK, map[string]any{
		"content":  dtos,
		"metadata": metadata,
	}, nil, h.errHandler)
}

func (h *TransactionHandler) FindById(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.FindById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	model, err := h.service.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch transaction")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.Update")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	var dto TransactionDTO
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
	span.SetAttributes(attribute.String("transaction.id", id.String()))

	if err := h.service.Update(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update transaction")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *TransactionHandler) DeleteById(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.DeleteById")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("transaction.id", id.String()))

	if err := h.service.DeleteById(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete transaction")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}

func (h *TransactionHandler) DeleteByCategoryId(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.DeleteByCategoryId")
	defer span.End()

	id, ok := handler.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("transaction.id", id.String()))

	if err := h.service.DeleteByCategoryId(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete transaction")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}

func parseTransactionQuery(r *http.Request) (transactionQuery, error) {
	v := validator.New()

	f := filters.Filters{
		Page:         httputil.ReadIntParam(r, "page", 1, v),
		PageSize:     httputil.ReadIntParam(r, "page_size", 20, v),
		Sort:         httputil.ReadStringParam(r, "sort", "id"),
		SortSafelist: []string{"id", "amount", "-id", "-amount"},
	}

	if filters.ValidateFilters(v, f); !v.Valid() {
		return transactionQuery{}, apiError.NewValidationError(v.Errors)
	}

	query := transactionQuery{
		Filters:    f,
		StartDate:  httputil.ReadDateParam(r, "start_date", v),
		EndDate:    httputil.ReadDateParam(r, "end_date", v),
		CategoryID: httputil.ReadUUIDParam(r, "category_id", v)}

	typeStr := httputil.ReadStringParam(r, "type", "")
	if typeStr != "" {
		ct, err := parseCategoryType(typeStr)
		if err != nil {
			return transactionQuery{}, apiError.NewValidationError(map[string]string{
				"type": "must be 'entrada' or 'saida'",
			})
		}
		query.Type = &ct
	}

	return query, nil
}

func (h *TransactionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_transaction/internal/features/transactions")
	ctx, span := tracer.Start(r.Context(), "TransactionHandler.Summary")
	defer span.End()

	query, err := parseSummaryQuery(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid query parameters")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	summary, err := h.service.Summary(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to generate summary")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	handler.Respond(w, r, http.StatusOK, summary, nil, h.errHandler)
}

func parseSummaryQuery(r *http.Request) (SummaryQuery, error) {
	v := validator.New()

	q := SummaryQuery{
		StartDate: httputil.ReadDateParam(r, "start_date", v),
		EndDate:   httputil.ReadDateParam(r, "end_date", v),
	}

	catStr := httputil.ReadStringParam(r, "category_id", "")
	if catStr != "" {
		id, err := uuid.Parse(catStr)
		if err != nil {
			v.AddError("category_id", "must be a valid uuid")
		} else {
			q.CategoryID = &id
		}
	}

	typeStr := httputil.ReadStringParam(r, "type", "")
	if typeStr != "" {
		ct, err := parseCategoryType(typeStr)
		if err != nil {
			v.AddError("type", "must be 'input' or 'output'")
		} else {
			q.Type = &ct
		}
	}

	if !v.Valid() {
		return SummaryQuery{}, apiError.NewValidationError(v.Errors)
	}

	return q, nil
}
