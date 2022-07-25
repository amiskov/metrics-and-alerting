package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Service interface {
	GetMetrics() []models.Metrics
}

type api struct {
	service Service
}

func New(s Service) *api {
	return &api{service: s}
}

func (a *api) Run(ctx context.Context, done chan bool, reportInterval time.Duration, serverURL string) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Println("Metrics report stopped.")
			done <- true
		default:
			a.sendMetrics(serverURL)
		}
	}
}

func (a *api) sendMetrics(sendURL string) {
	var wg sync.WaitGroup

	for _, m := range a.service.GetMetrics() {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()

			var val string
			switch m.MType {
			case "gauge":
				val = strconv.FormatFloat(float64(*m.Value), 'f', 3, 64)
			case "counter":
				fmt.Printf("%v (%v): %+v\n", m.ID, m.MType, m.Value)
				val = strconv.FormatInt(int64(*m.Delta), 10)
			default:
				log.Printf("Unknown metric type: %#v", m)
			}
			sendMetric(sendURL, m.MType, m.ID, val)
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
