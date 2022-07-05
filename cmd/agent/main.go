package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/sender"
	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

// TODO: Allow user to customize server URL e.g. with CLI flags.
var sendUrl = "http://127.0.0.1:8080"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	m := metrics.Metrics{}

	pollInterval := time.Duration(2 * time.Second)
	ticker := time.NewTicker(pollInterval)
	reportInterval := time.Duration(10 * time.Second)
	startTime := time.Now()

	for {
		select {
		case t := <-ticker.C:
			m.Update()
			elapsedFromStart := t.Sub(startTime).Round(time.Second)
			if elapsedFromStart%reportInterval == 0 {
				go func() {
					// TODO: Should we use WaitGroup or something?
					sender.SendMetrics(sendUrl, m)
				}()
			}
		case <-ctx.Done():
			ticker.Stop()
			// TODO: Stop all other things
			stop()
			log.Println("Signal received.")
			os.Exit(0)
		}
	}
}
