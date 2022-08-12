package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

type config struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	HashingKey    string
	PgDSN         string
}

func Parse() *config {
	cfg := config{
		// Defaults
		Address:       "localhost:8080",
		Restore:       true,
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	cfg.updateFromFlags()
	cfg.updateFromEnv()
	return &cfg
}

func (cfg *config) updateFromFlags() {
	flagAddress := flag.String("a", cfg.Address, "Server address.")
	flagRestore := flag.Bool("r", cfg.Restore, "Should server restore metrics from file on start?")
	flagStoreInterval := flag.Duration("i", cfg.StoreInterval, "Report interval in seconds.")
	flagStoreFile := flag.String("f", cfg.StoreFile, "File to store metrics.")
	flagHashingKey := flag.String("k", cfg.HashingKey, "Hashing key.")
	flagPgDSN := flag.String("d", cfg.PgDSN, "Postgres DSN.")

	flag.Parse()

	cfg.Address = *flagAddress
	cfg.Restore = *flagRestore
	cfg.StoreInterval = *flagStoreInterval
	cfg.StoreFile = *flagStoreFile
	cfg.HashingKey = *flagHashingKey
	cfg.PgDSN = *flagPgDSN // priority is higher than `flagStoreFile`
}

func (cfg *config) updateFromEnv() {
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
	if hashingKey, ok := os.LookupEnv("KEY"); ok {
		cfg.HashingKey = hashingKey
	}
	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.PgDSN = dsn
	}
}
