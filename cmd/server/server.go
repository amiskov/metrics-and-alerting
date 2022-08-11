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
	envCfg := config{
		// Defaults
		Address:       "localhost:8080",
		Restore:       true,
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	envCfg.UpdateFromFlags()
	envCfg.UpdateFromEnv()

	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan bool)

	storeCfg := store.Cfg{
		StoreFile:     envCfg.StoreFile,
		StoreInterval: envCfg.StoreInterval,
		Restore:       envCfg.Restore,
		Ctx:           ctx,
		Finished:      finished, // to make sure we wrote the data while terminating
	}

	storage, closeFile, err := store.New(&storeCfg)
	if err != nil {
		log.Fatalln("Can't init server store:", err)
	}
	defer closeFile()
	log.Printf("Server store created with config: %+v", envCfg)

	metricsAPI := api.New(storage)
	go metricsAPI.Run(envCfg.Address)
	log.Printf("Serving at http://%s\n", envCfg.Address)

	// Managing user signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-osSignalCtx.Done()
	log.Println("Terminating server, please wait...")
	cancel()
	stopBySyscall()

	<-finished
	close(finished)
	log.Println("Server has been successfully terminated. Bye!")
}

func (cfg *config) UpdateFromFlags() {
	flagAddress := flag.String("a", cfg.Address, "Server address.")
	flagRestore := flag.Bool("r", cfg.Restore, "Should server restore metrics from file on start?")
	flagStoreInterval := flag.Duration("i", cfg.StoreInterval, "Report interval in seconds.")
	flagStoreFile := flag.String("f", cfg.StoreFile, "File to store metrics.")

	flag.Parse()

	cfg.Address = *flagAddress
	cfg.Restore = *flagRestore
	cfg.StoreInterval = *flagStoreInterval
	cfg.StoreFile = *flagStoreFile
}

func (cfg *config) UpdateFromEnv() {
	if addr, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Address = addr
	}
	if file, ok := os.LookupEnv("STORE_FILE"); ok {
		cfg.StoreFile = file
	}
	if restoreEnv, ok := os.LookupEnv("RESTORE"); ok {
		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			log.Fatalf("Can't parse %s env var: %s", restoreEnv, err.Error())
		}
		cfg.Restore = restore
	}
	if intervalEnv, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		storeInterval, err := time.ParseDuration(intervalEnv)
		if err != nil {
			log.Fatalf("Can't parse %s env var: %s", intervalEnv, err.Error())
		}
		cfg.StoreInterval = storeInterval
	}
}
