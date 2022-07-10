package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Service interface {
	GetAllMetrics() []models.MetricRaw
}

type api struct {
	service Service
}

func New(s Service) *api {
	return &api{service: s}
}

func (api *api) SendMetrics(sendURL string) {
	var wg sync.WaitGroup

	for _, m := range api.service.GetAllMetrics() {
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
