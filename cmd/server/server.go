package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/backup"
	"github.com/amiskov/metrics-and-alerting/pkg/backup/file"
	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/server/api"
	"github.com/amiskov/metrics-and-alerting/pkg/server/repo"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/inmem"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	envCfg := config.Parse()

	lggr := logger.Run(envCfg.LogLevel)

	terminated := make(chan bool)

	storage, closeStorage := initStorage(ctx, terminated, envCfg)
	defer closeStorage()

	repo := repo.New(ctx, []byte(envCfg.HashingKey), storage)

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
	cancel()
	stopBySyscall()

	<-terminated
	close(terminated)
	log.Println("Server has been successfully terminated. Bye!")
}

func initStorage(ctx context.Context, terminated chan bool, cfg *config.Config) (repo.Storage, func()) {
	// Using PostgreSQL
	if cfg.PgDSN != "" {
		db, closer := postgres.New(ctx, cfg)
		db.Migrate()

		go func() {
			<-ctx.Done()
			// PostgreSQL will be terminated with `closer`.
			terminated <- true
		}()

		return db, closer
	}

	inmemory := inmem.New(ctx, []byte(cfg.HashingKey))
	closer := func() {}

	if cfg.StoreFile != "" {
		fileStorage, closeFile, err := file.New(cfg.StoreFile)
		if err != nil {
			logger.Log(ctx).Errorf("failed creating file storage: %s", err.Error())
		}

		backup := backup.New(ctx, terminated, inmemory, fileStorage)
		go backup.Run(cfg.Restore, cfg.StoreInterval)

		closer = func() {
			if err := closeFile(); err != nil {
				logger.Log(ctx).Errorf("failed closing file `%s`", cfg.StoreFile)
			}
		}
	}

	return inmemory, closer
}
