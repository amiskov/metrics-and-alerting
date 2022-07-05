package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/sender"
	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

func main() {
	// OS signals
	osSignalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// CLI options
	sendProtocol := flag.String("protocol", "http", "server protocol")
	sendHost := flag.String("host", "127.0.0.1", "server host")
	sendPort := flag.Int("port", 8080, "server port")
	pollIntervalNumber := flag.Int("poll", 2, "poll interval in seconds")
	reportIntervalNumber := flag.Int("report", 10, "report interval in seconds")
	flag.Parse()

	var sendURL = *sendProtocol + "://" + *sendHost + ":" + strconv.Itoa(*sendPort)

	m := metrics.Metrics{}

	pollInterval := time.Duration(time.Duration(*pollIntervalNumber) * time.Second)
	reportInterval := time.Duration(time.Duration(*reportIntervalNumber) * time.Second)
	ticker := time.NewTicker(pollInterval)
	startTime := time.Now()

	log.Printf("Agent started. Sending to: %v. Poll: %v. Report: %v.\n", sendURL, pollInterval, reportInterval)

	for {
		select {
		case t := <-ticker.C:
			m.Update()
			elapsedFromStart := t.Sub(startTime).Round(time.Second)
			if elapsedFromStart%reportInterval == 0 {
				go func() {
					// TODO: Should we use WaitGroup or something?
					sender.SendMetrics(sendURL, m)
				}()
			}
		case <-osSignalCtx.Done():
			ticker.Stop()
			// TODO: Stop all other things
			stop()
			os.Exit(0)
		}
	}
}
