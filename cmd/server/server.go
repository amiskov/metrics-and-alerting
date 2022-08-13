package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/cmd/server/db"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/file"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan bool)

	envCfg := config.Parse()

	storage, storageCloser := initStorage(ctx, finished, envCfg)
	defer storageCloser()

	repo := repo.New(envCfg, storage)

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

func initStorage(ctx context.Context, finished chan bool, envCfg *config.Config) (repo.Storage, func()) {
	if envCfg.PgDSN == "" {
		storeCfg := file.Cfg{
			StoreFile:     envCfg.StoreFile,
			StoreInterval: envCfg.StoreInterval,
			Restore:       envCfg.Restore,
			Ctx:           ctx,
			Finished:      finished, // to make sure we wrote the data while terminating
			HashingKey:    []byte(envCfg.HashingKey),
		}

		storage, closeFile, err := file.New(&storeCfg)
		if err != nil {
			log.Println("Can't init server store:", err)
		}
		closer := func() {
			if err := closeFile(); err != nil {
				log.Println("failed closing file storage:", err)
			}
		}

		log.Println("File is used for storing metrics.")

		return storage, closer
	}

	db, closer := db.New(ctx, envCfg)
	defer closer()

	// Always creates new `gauge` and `counter` PG types and `metrics` table.
	if dbErr := db.Migrate("sql/schema.sql"); dbErr != nil {
		log.Println(dbErr)
	}

	log.Println("PostgresDB is used for storing metrics.")

	return db, closer
}
