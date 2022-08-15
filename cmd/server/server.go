package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/db"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/file"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/inmem"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/intervaldump"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	envCfg := config.Parse()

	// TODO: move to the case with inmem db and file storage
	finished := make(chan bool)

	storage, closeStorage := initStorage(ctx, finished, envCfg)
	defer closeStorage()

	repo := repo.New(ctx, envCfg, storage)

	metricsAPI := api.New(repo)
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

func initStorage(ctx context.Context, finished chan bool, cfg *config.Config) (repo.Storage, func()) {
	// Using PostgreSQL
	if cfg.PgDSN != "" {
		db, closer := db.New(ctx, cfg)
		db.Migrate()

		go func() {
			<-ctx.Done()
			finished <- true
		}()

		return db, closer
	}

	inmemory := inmem.New(ctx, []byte(cfg.HashingKey))

	if cfg.StoreFile != "" {
		fileStorage, closeFile, err := file.New(cfg.StoreFile)
		if err != nil {
			log.Println("failed creating file storage", err)
		}

		dumper := intervaldump.New(ctx, finished, inmemory, fileStorage, cfg)
		go dumper.Run(cfg.Restore, cfg.StoreInterval)

		return inmemory, func() {
			if err := closeFile(); err != nil {
				log.Println("failed closing file", cfg.StoreFile)
			}
		}
	}

	return inmemory, func() {}
}
