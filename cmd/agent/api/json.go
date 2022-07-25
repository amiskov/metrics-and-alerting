package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

func (a *api) sendMetricsJSON(sendURL string) {
	var wg sync.WaitGroup

	for _, m := range a.service.GetMetrics() {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetricJSON(sendURL, m)
		}()
	}

	wg.Wait()
}

func sendMetricJSON(sendURL string, m models.Metrics) {
	postURL := sendURL + "/update/"
	contentType := "Content-Type: application/json"

	client := http.Client{}
	client.Timeout = 10 * time.Second

	jbz, err := json.Marshal(m)
	if err != nil {
		log.Printf("Error marshalling JSON: %+v", err)
	}

	resp, errPost := client.Post(postURL, contentType, bytes.NewBuffer(jbz))
	if errPost != nil {
		log.Println(errPost)
		return
	}
	defer resp.Body.Close()

	log.Printf("Sent JSON %+v to `%s`.\n", string(jbz), postURL)
}
