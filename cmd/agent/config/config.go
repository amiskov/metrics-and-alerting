package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type config struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
	HashingKey     string
	LogLevel       string
}

func NewConfig() *config {
	cfg := config{
		// Defaults
		Address:        "localhost:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
		LogLevel:       "warn",
	}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("config parsing failed:", err)
	}
	cfg.updateFromFlags()
	cfg.updateFromEnv()

	return &cfg
}

func (cfg *config) updateFromFlags() {
	flagAddress := flag.String("a", cfg.Address, "Server address.")
	flagReportInterval := flag.Duration("r", cfg.ReportInterval, "Report interval in seconds.")
	flagPollInterval := flag.Duration("p", cfg.PollInterval, "Poll interval in seconds.")
	flagHash := flag.String("k", cfg.HashingKey, "Hashing key.")
	flagLogLevel := flag.String("ll", cfg.LogLevel, "Logging Level.")

	flag.Parse()

	cfg.Address = *flagAddress
	cfg.ReportInterval = *flagReportInterval
	cfg.PollInterval = *flagPollInterval
	cfg.HashingKey = *flagHash
	cfg.LogLevel = *flagLogLevel
}

func (cfg *config) updateFromEnv() {
	if addr, ok := os.LookupEnv("ADDRESS"); ok {
		cfg.Address = addr
	}
	if dur, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollInterval, err := time.ParseDuration(dur)
		if err != nil {
			log.Fatalf("Can't parse %s: %s", dur, err.Error())
		}
		cfg.PollInterval = pollInterval
	}
	if dur, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportInterval, err := time.ParseDuration(dur)
		if err != nil {
			log.Fatalf("Can't parse %s: %s", dur, err.Error())
		}
		cfg.ReportInterval = reportInterval
	}
	if hashingKey, ok := os.LookupEnv("KEY"); ok {
		cfg.HashingKey = hashingKey
	}
	if ll, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = ll
	}
}
