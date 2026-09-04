package api

import (
	"ms_transaction/internal/core/config"
	"ms_transaction/internal/core/httpclient"
	"ms_transaction/internal/core/httpclient/categories"
)

type clients struct {
	category *categories.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		category: categories.NewClient(
			httpclient.Config{
				BaseURL: cfg.Clients.CategoryURL,
				Timeout: cfg.Server.Timeout,
			},
		),
	}
}
