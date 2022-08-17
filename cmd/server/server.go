package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/server/api"
	"github.com/amiskov/metrics-and-alerting/pkg/server/repo"
)

func main() {
	appCtx, cancelAppCtx := context.WithCancel(context.Background())
	envCfg := config.Parse()

	lggr := logger.Run(envCfg.LogLevel)

	storage, closeStorage := initStorage(appCtx, envCfg)
	defer closeStorage()

	repo := repo.New(appCtx, []byte(envCfg.HashingKey), storage)

	if envCfg.StoreFile != "" {
		terminated := make(chan bool)
		initBackupToFile(appCtx, storage, terminated, envCfg)
		go func() {
			<-terminated
			close(terminated)
			log.Println("Data was successfully backed up.")
		}()
	}

	metricsAPI := api.New(repo, logger.NewLoggingMiddleware(lggr))
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
	cancelAppCtx()
	stopBySyscall()
}
