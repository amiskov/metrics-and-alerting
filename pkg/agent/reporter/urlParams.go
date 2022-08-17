package reporter

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

func (r *reporter) sendMetrics(metrics []models.Metrics) {
	var wg sync.WaitGroup
	for _, m := range metrics {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.sendMetric(m)
		}()
	}
	wg.Wait()
}

func (r reporter) sendMetric(m models.Metrics) {
	val, err := m.GetStrVal()
	if err != nil {
		logger.Log(r.ctx).Error("bad metric format: %#v", m)
		return
	}
	postURL := r.serverURL + "/update/" + m.MType + "/" + m.ID + "/" + val
	client := http.Client{}
	client.Timeout = 10 * time.Second
	resp, errPost := client.Post(postURL, "Content-Type: text/plain", nil)
	if errPost != nil {
		log.Printf("Failed to send metric. URL: `%s`. Error: %v\n", postURL, errPost)
		return
	}
	defer resp.Body.Close()
	log.Printf("Sent to `%s`.\n", postURL)
}
