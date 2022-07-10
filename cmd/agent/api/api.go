package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	UpdateAll()
	GetAll() []models.MetricRaw
}

func SendMetrics(sendURL string, metrics []models.MetricRaw) {
	var wg sync.WaitGroup

	for _, m := range metrics {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetric(sendURL, m.Type, m.Name, m.Value)
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
