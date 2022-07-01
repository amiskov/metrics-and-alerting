package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultHost = "http://localhost"
const defaultPort = ":8080"
const contentType = "Content-Type: text/plain"

// Returns a URL to send a metric.
// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
func prepareURL(t string, n string, v string) string {
	return defaultHost + defaultPort + "/update/" + t + "/" + n + "/" + v
}

func sendMetric(mType string, mName string, mValue string) {
	url := prepareURL(mType, mName, mValue)
	contentType := "Content-Type: text/plain"
	client := http.Client{}
	client.Timeout = 10 * time.Second
	resp, errPost := client.Post(url, contentType, nil)
	if errPost != nil {
		panic(errPost)
	}
	r, errRespBody := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if errRespBody != nil {
		panic(errRespBody)
	}
	fmt.Println("Sent! Server said:", string(r))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	ticker := time.NewTicker(500 * time.Millisecond)
	c := 0

	for {
		select {
		case t := <-ticker.C:
			metricType := "gauge"
			metricName := "Alloc"
			metricValue := "11111"
			sendMetric(metricType, metricName, metricValue)
			fmt.Println(t.Second())
			c++
			if c >= 10 {
				ticker.Stop()
				return
			}
		case <-ctx.Done():
			ticker.Stop()
			// TODO: Stop all other things
			stop()
			fmt.Println("Signal received.")
			os.Exit(0)
		}
	}
}
