package transactions

import (
	"context"
	"ms_category/internal/core/domain/apiError"
	"ms_category/internal/core/httpclient"
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

func (c *HTTPClient) DeleteByCategoryId(
	ctx context.Context,
	id uuid.UUID,
) error {
	return c.base.Delete(
		ctx,
		"/category/%s"+id.String(),
		nil,
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
}
