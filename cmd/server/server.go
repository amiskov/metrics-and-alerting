package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
	"github.com/caarlos0/env"
)

const (
	defaultAddress       = "localhost:8080"
	defaultRestore       = true
	defaultStoreInterval = time.Duration(300 * time.Second)
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
)

type config struct {
	Address string `env:"ADDRESS" envDefault:"nope"`
	// 0s STORE_INTERVAL stores metrics immediately.
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"-1s"`
	// Don't store metrics if `StoreFile` is empty.
	StoreFile string `env:"STORE_FILE" envDefault:"nope"`
	// Load previously saved metrics from file if `Restore` is true.
	Restore bool `env:"RESTORE" envDefault:"false"`
}

func (cfg *config) UpdateFromCLI() {
	cliAddress := flag.String("a", defaultAddress, "Server address")
	cliRestore := flag.Bool("a", defaultRestore, "Should server restore on start?")
	cliStoreInterval := flag.Duration("i", defaultStoreInterval, "Report interval in seconds.")
	cliStoreFile := flag.String("f", defaultStoreFile, "Server address.")

	flag.Parse()

	notSetDuration := time.Duration(-1 * time.Second)

	if cfg.Address == "nope" || *cliAddress != defaultAddress {
		cfg.Address = *cliAddress
	}
	if cfg.Restore == false || *cliRestore != defaultRestore {
		cfg.Restore = *cliRestore
	}
	if cfg.StoreInterval == notSetDuration || *cliStoreInterval != defaultStoreInterval {
		cfg.StoreInterval = *cliStoreInterval
	}
	if cfg.StoreFile == "nope" || *cliStoreFile != defaultStoreFile {
		cfg.StoreFile = *cliStoreFile
	}
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("Parsing CLI params failed.", err)
	}
	cfg.UpdateFromCLI()

	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan bool)

	storage, err := store.New(store.StoreCfg{
		StoreFile:     cfg.StoreFile,
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
		Ctx:           ctx,
		Finished:      finished,
	})
	if err != nil {
		log.Fatalln("Creating server store failed.", err)
	}
	defer storage.CloseFile()

	metricsAPI := api.New(storage)
	go metricsAPI.Run(cfg.Address)

	// Managing user signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	<-osSignalCtx.Done()
	fmt.Println("Terminating server, please wait...")
	cancel()
	stopBySyscall()

	<-finished
	close(finished)
	fmt.Println("Server has been successfully terminated. Bye!")
	os.Exit(0)
}
