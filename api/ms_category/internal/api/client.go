package api

import (
	"ms_category/internal/core/config"
	"ms_category/internal/core/httpclient/transactions"
	"shared/httpx/httpclient"
)

type clients struct {
	transactions *transactions.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		transactions: transactions.NewClient(
			httpclient.Config{
				BaseURL: cfg.Clients.TransactionURL,
				Timeout: cfg.Base.Server.Timeout,
			},
		),
	}
}
