package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/api"
	"github.com/amiskov/metrics-and-alerting/cmd/agent/service"
	"github.com/caarlos0/env"
)

type config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	updater := service.New()
	finished := make(chan bool, 1) // buffer of 2 for updater and reporter
	go updater.Run(ctx, finished, cfg.PollInterval)

	reporter := api.New(updater, ctx, finished, cfg.ReportInterval, cfg.Address)
	// go reporter.ReportWithURLParams()
	go reporter.ReportWithJSON()

	log.Println("Agent has been started.")
	log.Printf("Sending to: %v. Poll: %v. Report: %v.\n", cfg.Address,
		cfg.PollInterval, cfg.ReportInterval)

	// Managing user signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-osSignalCtx.Done()
	fmt.Println("Terminating agent, please wait...")
	cancel() // stop processes
	stopBySyscall()

	<-finished
	<-finished
	close(finished)

	fmt.Println("Agent has been terminated. Bye!")
	os.Exit(0)
}
