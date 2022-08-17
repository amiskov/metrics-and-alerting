package main

import (
	"context"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/backup"
	"github.com/amiskov/metrics-and-alerting/pkg/backup/filestore"
	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/server/repo"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/inmem"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/postgres"
)

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
		logger.Log(ctx).Errorf("failed creating file storage: %s", err.Error())
		return nil
	}

	backup := backup.New(ctx, terminated, db, storeToBackup)
	go backup.Run(cfg.Restore, cfg.StoreInterval)

	return func() {
		if err := closeFile(); err != nil {
			logger.Log(ctx).Errorf("failed closing file `%s`", cfg.StoreFile)
		}
	}
}
