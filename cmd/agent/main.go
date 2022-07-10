package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/api"
	"github.com/amiskov/metrics-and-alerting/cmd/agent/store"
)

var serverURL string
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

	serverURL = *sendProtocol + "://" + *sendHost + ":" + strconv.Itoa(*sendPort)

	pollInterval = time.Duration(time.Duration(*pollIntervalNumber) * time.Second)
	reportInterval = time.Duration(time.Duration(*reportIntervalNumber) * time.Second)
}

func main() {
	// OS signals
	osSignalCtx, stopBySyscall := signal.NotifyContext(
		context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stopBySyscall()

	metricsStore := store.NewMetricsStore()

	ticker := time.NewTicker(pollInterval)
	startTime := time.Now()

	log.Printf("Agent started. Sending to: %v. Poll: %v. Report: %v.\n",
		serverURL, pollInterval, reportInterval)

	var wg sync.WaitGroup

	for {
		select {
		case t := <-ticker.C:
			metricsStore.UpdateAll()
			elapsedFromStart := t.Sub(startTime).Round(time.Second)
			if elapsedFromStart%reportInterval == 0 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					api.SendMetrics(serverURL, metricsStore.GetAll())
				}()
			}
		case <-osSignalCtx.Done():
			ticker.Stop()
			// TODO: What else should we stop after receiving the terminating OS signal?
			stopBySyscall()
			os.Exit(0)
		}

		wg.Wait()
	}
}
