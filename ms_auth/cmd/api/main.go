package main

import (
	"ms_auth/internal/api"
	"ms_auth/internal/core/config"
)

func main() {
	cfg := config.New()
	cfg.Env = "development"
	app := api.NewApp(*cfg)
	err := app.Server()
	if err != nil {
		app.Logger.Error(err.Error())
	}
}
