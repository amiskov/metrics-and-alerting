package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/metric"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	m := &metric.Metrics{}
	pollInterval := time.Duration(2 * time.Second)
	reportInterval := time.Duration(10 * time.Second)
	startTime := time.Now()

	ticker := time.NewTicker(pollInterval)

	for {
		select {
		case t := <-ticker.C:
			m.Update()
			elapsedFromStart := t.Sub(startTime).Round(time.Second)
			if elapsedFromStart%reportInterval == 0 {
				go func() {
					// TODO: Should we use WaitGroup or something?
					m.Send()
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
