package categories

import (
	"context"
	"ms_transaction/internal/core/domain/apiError"
	"ms_transaction/internal/core/httpclient"
	"net/http"
	"uuid"
)

type HTTPClient struct {
	base *httpclient.BaseClient
}

func NewClient(cfg httpclient.Config) *HTTPClient {
	return &HTTPClient{
		base: httpclient.New(cfg.BaseURL, cfg.Timeout),
	}
}

type CategoryDTO struct {
	ID     *uuid.UUID `json:"id,omitempty"`
	UserID *uuid.UUID `json:"user_id,omitempty"`
	Name   *string    `json:"name,omitempty"`
	Type   *string    `json:"type"`
	GoalID *uuid.UUID `json:"goal_id,omitempty"`
}

func (c *HTTPClient) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (CategoryDTO, error) {
	var response CategoryDTO

	err := c.base.Get(
		ctx,
		"/"+id.String(),
		&response,
		func(status int, body []byte) error {
			if status == http.StatusNotFound {
				return apiError.NewHTTPError(
					"category not found",
					http.StatusNotFound,
					nil,
				)
			}
			return nil
		},
	)

	if err != nil {
		return CategoryDTO{}, err
	}

	return response, nil
}
