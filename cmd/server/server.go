package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
)

type config struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

func main() {
	cfg := config{
		// Config Defaults
		Address:       "localhost:8080",
		Restore:       true,
		StoreInterval: time.Duration(300 * time.Second),
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	cfg.UpdateFromCLI()
	cfg.UpdateFromEnv()
	log.Printf("Config is: %#v", cfg)

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
	log.Println("Terminating server, please wait...")
	cancel()
	stopBySyscall()

	<-finished
	close(finished)
	log.Println("Server has been successfully terminated. Bye!")
	os.Exit(0)
}

func (cfg *config) UpdateFromCLI() {
	cliAddress := flag.String("a", cfg.Address, "Server address.")
	cliRestore := flag.Bool("r", cfg.Restore, "Should server restore metrics from file on start?")
	cliStoreInterval := flag.Duration("i", cfg.StoreInterval, "Report interval in seconds.")
	cliStoreFile := flag.String("f", cfg.StoreFile, "File to store metrics.")

	flag.Parse()

	cfg.Address = *cliAddress
	cfg.Restore = *cliRestore
	cfg.StoreInterval = *cliStoreInterval
	cfg.StoreFile = *cliStoreFile
}

func (cfg *config) UpdateFromEnv() {
	if addr := os.Getenv("ADDRESS"); addr != "" {
		cfg.Address = addr
	}
	if f := os.Getenv("STORE_FILE"); f != "" {
		cfg.StoreFile = f
	}
	if r := os.Getenv("RESTORE"); r != "" {
		restore, err := strconv.ParseBool(r)
		if err != nil {
			cfg.Restore = restore
		}
	}
	if dur := os.Getenv("STORE_INTERVAL"); dur != "" {
		storeInterval, err := time.ParseDuration(dur)
		if err != nil {
			log.Fatalf("Can't parse %s: %s", dur, err.Error())
		}
		cfg.StoreInterval = storeInterval
	}
}
