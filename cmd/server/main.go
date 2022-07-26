package main

import (
	"fmt"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
	"github.com/caarlos0/env"
)

type config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	storage := store.NewServerStore()

	metricsAPI := api.New(storage)
	metricsAPI.Run(cfg.Address)
}
