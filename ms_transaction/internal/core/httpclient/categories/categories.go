package categories

import (
	"context"
	"fmt"
	"net/http"
	"shared/auth/domain/apiError"
	"shared/httpx/httpclient"
	"shared/utils/filters"
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
	ID     uuid.UUID  `json:"id,omitempty"`
	UserID uuid.UUID  `json:"user_id,omitempty"`
	Name   string     `json:"name,omitempty"`
	Type   string     `json:"type"`
	GoalID *uuid.UUID `json:"goal_id,omitempty"`
}

type CategoryListResponse struct {
	Content  []CategoryDTO    `json:"content"`
	Metadata filters.Metadata `json:"metadata"`
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

func (c *HTTPClient) FindAll(
	ctx context.Context,
	page int,
	pageSize int,
) ([]CategoryDTO, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	path := fmt.Sprintf("/?page=%d&page_size=%d", page, pageSize)

	var response CategoryListResponse
	err := c.base.Get(ctx, path, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.Content, nil
}
