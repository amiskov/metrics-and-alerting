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

var sendURL string
var pollInterval time.Duration
var reportInterval time.Duration

func init() {
	// CLI options
	sendProtocol := flag.String("protocol", "http", "server protocol")
	sendHost := flag.String("host", "127.0.0.1", "server host")
	sendPort := flag.Int("port", 8080, "server port")
	pollIntervalNumber := flag.Int("poll", 2, "poll interval in seconds")
	reportIntervalNumber := flag.Int("report", 10, "report interval in seconds")
	flag.Parse()

	sendURL = *sendProtocol + "://" + *sendHost + ":" + strconv.Itoa(*sendPort)

	pollInterval = time.Duration(time.Duration(*pollIntervalNumber) * time.Second)
	reportInterval = time.Duration(time.Duration(*reportIntervalNumber) * time.Second)
}

func main() {
	// OS signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stopBySyscall()

	m := metrics.Metrics{}

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
			stopBySyscall()
			os.Exit(0)
		}
	}
}
