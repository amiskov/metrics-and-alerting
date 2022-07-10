package sender

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/model"
)

type Storer interface {
	UpdateAll()
	GetAll() []model.MetricForSend
}

func SendMetrics(sendURL string, metrics []model.MetricForSend) {
	var wg sync.WaitGroup

	for _, metric := range metrics {
		metric := metric
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetric(sendURL, metric.Type, metric.Name, metric.Value)
		}()
	}

	wg.Wait()
}

func sendMetric(sendURL string, mType string, mName string, mValue string) {
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
