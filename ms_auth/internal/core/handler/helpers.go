package handler

import (
	"fmt"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/filters"
	"ms_auth/internal/core/validator"
	"ms_auth/pkg/httpjson"
	"ms_auth/pkg/httputil"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type errorHandler interface {
	ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error)
	BadRequestResponse(w http.ResponseWriter, r *http.Request, err error)
}

func ParseIntID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errorHandler,
) (int64, bool) {
	id, err := readIntPathVariable(r, "id")
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return 0, false
	}
	return id, true
}

func GetFilters(r *http.Request, sortSafelist []string) (filters.Filters, error) {
	v := validator.New()

	var f = filters.Filters{}
	f.Page = httputil.ReadIntParam(r, "page", 1, v)
	f.PageSize = httputil.ReadIntParam(r, "page_size", 20, v)
	f.Sort = httputil.ReadStringParam(r, "sort", "")
	f.SortSafelist = sortSafelist

	if filters.ValidateFilters(v, f); !v.Valid() {
		return filters.Filters{}, apiError.NewValidationError(v.Errors)
	}

	return f, nil

}

func ParseStringField(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errorHandler,
	field string,
) (string, bool) {
	value, err := readStringPathVariable(r, field)
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return "", false
	}
	return value, true
}

func ParseUUID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp errorHandler,
) (uuid.UUID, bool) {

	id, err := readStringPathVariable(r, "id")
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return uuid.Nil, false
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		errRsp.BadRequestResponse(w, r, err)
		return uuid.Nil, false
	}

	return uid, true
}

func Respond(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
	headers http.Header,
	errRsp errorHandler,
) {
	err := httpjson.WriteJSON(w, status, data, headers)
	if err != nil {
		errRsp.ServerErrorResponse(w, r, err)
		return
	}
}

func readIntPathVariable(r *http.Request, key string) (int64, error) {
	s := r.PathValue(key)

	if s == "" {
		return 0, fmt.Errorf("missing path parameter: %s", key)
	}

	value, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}

	return value, nil
}

func readStringPathVariable(r *http.Request, key string) (string, error) {
	s := r.PathValue(key)
	if s == "" {
		return "", fmt.Errorf("missing path parameter: %s", key)
	}

	return s, nil
}
