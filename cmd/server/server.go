package main

import (
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
	"github.com/caarlos0/env"
)

type config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
	// 0s STORE_INTERVAL stores metrics immediately.
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	// Don't store metrics if `StoreFile` is empty.
	StoreFile string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	// Load previously saved metrics from file if `Restore` is true.
	Restore bool `env:"RESTORE" envDefault:"true"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("Parsing CLI params failed.", err)
	}

	storage, err := store.New(store.StoreCfg{
		StoreFile:     cfg.StoreFile,
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
	})
	if err != nil {
		log.Fatalln("Creating server store failed.", err)
	}
	defer storage.CloseFile()

	metricsAPI := api.New(storage)
	metricsAPI.Run(cfg.Address)
}
