package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/backup"
	"github.com/amiskov/metrics-and-alerting/pkg/backup/filestore"
	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/server/api"
	"github.com/amiskov/metrics-and-alerting/pkg/server/repo"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/inmem"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/postgres"
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

func initStorage(ctx context.Context, cfg *config.Config) (repo.Storage, func()) {
	// Using PostgreSQL
	if cfg.PgDSN != "" {
		db, closer := postgres.New(ctx, cfg)
		db.Migrate()
		return db, closer
	}

	db := inmem.New(ctx, []byte(cfg.HashingKey))
	return db, func() {}
}

func initBackupToFile(ctx context.Context, db backup.Sourcer, terminated chan bool, cfg *config.Config) func() {
	storeToBackup, closeFile, err := filestore.New(cfg.StoreFile)
	if err != nil {
		logger.Log(ctx).Errorf("main: failed creating file storage: %s", err.Error())
		return nil
	}

	backup := backup.New(ctx, terminated, db, storeToBackup)
	go backup.Run(cfg.Restore, cfg.StoreInterval)

	return func() {
		if err := closeFile(); err != nil {
			logger.Log(ctx).Errorf("main: failed closing file `%s`", cfg.StoreFile)
		}
	}
}
