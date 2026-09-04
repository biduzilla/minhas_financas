package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"ms_transaction/internal/core/contexts"
	"ms_transaction/internal/core/domain/apiError"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type BaseClient struct {
	baseURL    string
	httpClient *http.Client
}

type Config struct {
	BaseURL string
	Timeout time.Duration
}

func New(baseURL string, timeout time.Duration) *BaseClient {
	return &BaseClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (c *BaseClient) Post(ctx context.Context, path string, request any, response any, errorMapper func(status int, body []byte) error) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	return c.do(req, response, errorMapper)
}

func (c *BaseClient) Get(ctx context.Context, path string, response any, errorMapper func(status int, body []byte) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	return c.do(req, response, errorMapper)
}

func (c *BaseClient) Delete(ctx context.Context, path string, response any, errorMapper func(status int, body []byte) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	return c.do(req, response, errorMapper)
}

func (c *BaseClient) do(
	req *http.Request,
	response any,
	errorMapper func(status int, body []byte,
	) error) error {
	req.Header.Set("Content-Type", "application/json")

	token := contexts.GetToken(req.Context())
	if token == "" {
		return fmt.Errorf("internal error: missing authentication token in context")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		if errorMapper != nil {
			if mappedErr := errorMapper(resp.StatusCode, body); mappedErr != nil {
				return mappedErr
			}
		}
		return parseServiceError(body, resp.StatusCode)
	}

	if response != nil {
		if err := json.Unmarshal(body, response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func parseServiceError(body []byte, status int) error {
	var payload struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &payload); err == nil && payload.Message != "" {
		return apiError.NewHTTPError(payload.Message, status, errors.New(payload.Message))
	}

	return fmt.Errorf("service returned status %d: %s", status, string(body))
}
