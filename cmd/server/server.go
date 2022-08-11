package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan bool)

	envCfg := config.Parse()
	storeCfg := store.Cfg{
		StoreFile:     envCfg.StoreFile,
		StoreInterval: envCfg.StoreInterval,
		Restore:       envCfg.Restore,
		Ctx:           ctx,
		Finished:      finished, // to make sure we wrote the data while terminating
		HashingKey:    []byte(envCfg.HashingKey),
	}

	storage, closeFile, err := store.New(&storeCfg)
	if err != nil {
		log.Fatalln("Can't init server store:", err)
	}
	defer func() {
		if err := closeFile(); err != nil {
			log.Println("failed closing file storage:", err)
		}
	}()
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
