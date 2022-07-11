package main

import (
	"flag"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/agent/api"
	"github.com/amiskov/metrics-and-alerting/cmd/agent/service"
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
	metricsService := service.New()
	metricsAPI := api.New(metricsService)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		metricsService.Run(pollInterval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		metricsAPI.Run(reportInterval, serverURL)
	}()

	wg.Wait()

	log.Printf("Agent started. Sending to: %v. Poll: %v. Report: %v.\n",
		serverURL, pollInterval, reportInterval)
}
