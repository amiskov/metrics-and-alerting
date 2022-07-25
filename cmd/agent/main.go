package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	ctx, cancel := context.WithCancel(context.Background())

	updater := service.New()
	reporter := api.New(updater)

	finished := make(chan bool, 1) // buffer of 2 for updater and reporter
	// go updater.Run(ctx, finished, pollInterval)
	go reporter.RunJSON(ctx, finished, reportInterval, serverURL)

	log.Println("Agent has been started.")
	log.Printf("Sending to: %v. Poll: %v. Report: %v.\n", serverURL, pollInterval, reportInterval)

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
