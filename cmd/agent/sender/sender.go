package sender

import (
	"log"
	"net/http"
	"strconv"
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
			strVal := strconv.FormatFloat(float64(val), 'f', 2, 64)
			sendMetric(sendURL, "gauge", name, strVal)
		}()
	}

	// Sending PollCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		strVal := strconv.Itoa(int(m.PollCount))
		sendMetric(sendURL, "counter", "PollCount", strVal)
	}()

	// Sending RandomValue
	wg.Add(1)
	go func() {
		defer wg.Done()
		strVal := strconv.FormatFloat(float64(m.RandomValue), 'f', 2, 64)
		sendMetric(sendURL, "gauge", "RandomValue", strVal)
	}()

	wg.Wait()
}

func sendMetric(sendURL string, mType string, mName string, mValue string) {
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	postURL := sendURL + "/update/" + mType + "/" + mName + "/" + mValue
	contentType := "Content-Type: text/plain"
	client := http.Client{}
	client.Timeout = 10 * time.Second
	_, errPost := client.Post(postURL, contentType, nil)
	if errPost != nil {
		panic(errPost)
	}
	log.Printf("Sent to `%s`.\n", postURL)
}
