package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/config"
	"github.com/amiskov/metrics-and-alerting/pkg/agent/reporter"
	"github.com/amiskov/metrics-and-alerting/pkg/agent/updater"
	"github.com/amiskov/metrics-and-alerting/pkg/logger"
)

func main() {
	cfg := config.NewConfig()

	_ = logger.Run(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())
	terminated := make(chan bool, 1) // buffer of 2 for updater and reporter

	updater := updater.New([]byte(cfg.HashingKey))
	go updater.Run(ctx, terminated, cfg.PollInterval)

	reporter := reporter.New(ctx, updater, terminated, cfg.ReportInterval, cfg.Address, cfg.HashingKey)
	go reporter.ReportWithJSON()

	log.Printf("Agent started with config %+v\n.", cfg)
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
