package metric

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Gauge float64
type Counter int64

type Metrics struct {
	RuntimeMetrics map[string]Gauge
	PollCount      Counter
	RandomValue    Gauge
}

var memStats = &runtime.MemStats{}

func (m *Metrics) Update() {
	runtime.ReadMemStats(memStats)

	if m.RuntimeMetrics == nil {
		m.RuntimeMetrics = make(map[string]Gauge)
	}

	m.RuntimeMetrics["Alloc"] = Gauge(memStats.Alloc)
	m.RuntimeMetrics["BuckHashSys"] = Gauge(memStats.BuckHashSys)
	m.RuntimeMetrics["Frees"] = Gauge(memStats.Frees)
	m.RuntimeMetrics["GCCPUFraction"] = Gauge(memStats.GCCPUFraction)
	m.RuntimeMetrics["GCSys"] = Gauge(memStats.GCSys)
	m.RuntimeMetrics["HeapAlloc"] = Gauge(memStats.HeapAlloc)
	m.RuntimeMetrics["HeapIdle"] = Gauge(memStats.HeapIdle)
	m.RuntimeMetrics["HeapInuse"] = Gauge(memStats.HeapInuse)
	m.RuntimeMetrics["HeapObjects"] = Gauge(memStats.HeapObjects)
	m.RuntimeMetrics["HeapReleased"] = Gauge(memStats.HeapReleased)
	m.RuntimeMetrics["HeapSys"] = Gauge(memStats.HeapSys)
	m.RuntimeMetrics["LastGC"] = Gauge(memStats.LastGC)
	m.RuntimeMetrics["Lookups"] = Gauge(memStats.Lookups)
	m.RuntimeMetrics["MCacheInuse"] = Gauge(memStats.MCacheInuse)
	m.RuntimeMetrics["MCacheSys"] = Gauge(memStats.MCacheSys)
	m.RuntimeMetrics["MSpanInuse"] = Gauge(memStats.MSpanInuse)
	m.RuntimeMetrics["MSpanSys"] = Gauge(memStats.MSpanSys)
	m.RuntimeMetrics["Mallocs"] = Gauge(memStats.Mallocs)
	m.RuntimeMetrics["NextGC"] = Gauge(memStats.NextGC)
	m.RuntimeMetrics["NumForcedGC"] = Gauge(memStats.NumForcedGC)
	m.RuntimeMetrics["NumGC"] = Gauge(memStats.NumGC)
	m.RuntimeMetrics["OtherSys"] = Gauge(memStats.OtherSys)
	m.RuntimeMetrics["PauseTotalNs"] = Gauge(memStats.PauseTotalNs)
	m.RuntimeMetrics["StackInuse"] = Gauge(memStats.StackInuse)
	m.RuntimeMetrics["StackSys"] = Gauge(memStats.StackSys)
	m.RuntimeMetrics["Sys"] = Gauge(memStats.Sys)
	m.RuntimeMetrics["TotalAlloc"] = Gauge(memStats.TotalAlloc)

	m.PollCount++
	m.RandomValue = Gauge(rand.Float64())
}

func (m Metrics) Send() {
	var wg sync.WaitGroup

	// Sending Runtime Metrics
	for name, val := range m.RuntimeMetrics {
		name := name
		val := val

		wg.Add(1)
		go func() {
			defer wg.Done()
			strVal := strconv.FormatFloat(float64(val), 'f', 2, 64)
			sendMetric("gauge", name, strVal)
		}()
	}

	// Sending PollCount
	wg.Add(1)
	go func() {
		defer wg.Done()
		strVal := strconv.Itoa(int(m.PollCount))
		sendMetric("counter", "PollCount", strVal)
	}()

	// Sending RandomValue
	wg.Add(1)
	go func() {
		defer wg.Done()
		strVal := strconv.FormatFloat(float64(m.RandomValue), 'f', 2, 64)
		sendMetric("gauge", "RandomValue", strVal)
	}()

	wg.Wait()
}

func sendMetric(mType string, mName string, mValue string) {
	// Returns a URL to send a metric.
	// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
	url := "http://localhost" + ":8080" + "/update/" + mType + "/" + mName + "/" + mValue
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
	log.Println("Sent! Server said:", string(r))
}
