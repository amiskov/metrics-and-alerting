package main

import (
	"fmt"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store/inmemory"
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

	storage := inmemory.New()

	metricsAPI := api.New(storage)
	metricsAPI.Run(cfg.Address)
}
