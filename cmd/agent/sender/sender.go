package sender

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

func SendMetrics(sendURL string, m metrics.Metrics) {
	var wg sync.WaitGroup

	// Sending Runtime Metrics
	for name, val := range m.RuntimeMetrics {
		name := name
		val := val

		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetric(sendURL, "gauge", name, val.String())
		}()
	}

	// Sending PollCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		sendMetric(sendURL, "counter", "PollCount", m.PollCount.String())
	}()

	// Sending RandomValue
	wg.Add(1)
	go func() {
		defer wg.Done()
		sendMetric(sendURL, "gauge", "RandomValue", m.RandomValue.String())
	}()

	wg.Wait()
}

func sendMetric(sendURL string, mType string, mName string, mValue string) {
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	postURL := sendURL + "/update/" + mType + "/" + mName + "/" + mValue
	contentType := "Content-Type: text/plain"
	client := http.Client{}
	client.Timeout = 10 * time.Second
	resp, errPost := client.Post(postURL, contentType, nil)
	if errPost != nil {
		log.Println(errPost)
		return
	}
	defer resp.Body.Close()
	log.Printf("Sent to `%s`.\n", postURL)
}
