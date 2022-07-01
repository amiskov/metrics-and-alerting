package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultHost = "http://localhost"
const defaultPort = ":8080"
const contentType = "Content-Type: text/plain"

// Returns a URL to send a metric.
// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
func prepareUrl(t string, n string, v string) string {
	return defaultHost + defaultPort + "/update/" + t + "/" + n + "/" + v
}

func sendGaugeMetric(mType string, mName string, mValue string) {
	contentType := "Content-Type: text/plain"
	client := http.Client{}
	client.Timeout = 10 * time.Second
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	url := prepareUrl(mType, mName, mValue)
	resp, errPost := client.Post(url, contentType, nil)
	if errPost != nil {
		panic(errPost)
	}
	r, errRespBody := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if errRespBody != nil {
		panic(errRespBody)
	}
	fmt.Println(string(r))
}

func main() {
	// Агент должен штатно завершаться по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	sendGaugeMetric("gauge", "Alloc", "32323")
}
