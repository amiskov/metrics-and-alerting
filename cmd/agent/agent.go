package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/api"
	"github.com/amiskov/metrics-and-alerting/cmd/agent/service"
)

type config struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
	HashingKey     string
}

func main() {
	cfg := config{
		Address:        "localhost:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
	}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	cfg.updateFromFlags()
	cfg.updateFromEnv()
	log.Printf("Config is: %#v", cfg)

	ctx, cancel := context.WithCancel(context.Background())

	updater := service.New([]byte(cfg.HashingKey))
	terminated := make(chan bool, 1) // buffer of 2 for updater and reporter
	go updater.Run(ctx, terminated, cfg.PollInterval)

	reporter := api.New(ctx, updater, terminated, cfg.ReportInterval, cfg.Address)
	// go reporter.ReportWithURLParams()
	go reporter.ReportWithJSON()

	log.Println("Agent has been started.")
	log.Printf("Sending to: %v. Poll: %v. Report: %v.\n", cfg.Address,
		cfg.PollInterval, cfg.ReportInterval)

	// Managing user signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-osSignalCtx.Done()
	log.Println("Terminating agent, please wait...")
	cancel() // stop processes
	stopBySyscall()

	<-terminated
	<-terminated
	close(terminated)

	log.Println("Agent has been terminated. Bye!")
}

func (cfg *config) updateFromFlags() {
	flagAddress := flag.String("a", cfg.Address, "Server address.")
	flagReportInterval := flag.Duration("r", cfg.ReportInterval, "Report interval in seconds.")
	flagPollInterval := flag.Duration("p", cfg.PollInterval, "Poll interval in seconds.")
	flagHash := flag.String("k", cfg.HashingKey, "Hashing key.")

	flag.Parse()

	cfg.Address = *flagAddress
	cfg.ReportInterval = *flagReportInterval
	cfg.PollInterval = *flagPollInterval
	cfg.HashingKey = *flagHash
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
}
