// Package updater creates a snapshot of a running service with a bunch of metrics.
package updater

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/inmem"
)

type updater struct {
	mx         *sync.RWMutex
	ctx        context.Context
	memStats   *runtime.MemStats
	metrics    *inmem.DB
	hashingKey []byte
}

func New(ctx context.Context, key []byte) *updater {
	return &updater{
		mx:         new(sync.RWMutex),
		ctx:        ctx,
		memStats:   new(runtime.MemStats),
		metrics:    inmem.New(ctx, key),
		hashingKey: key,
	}
}

func (u *updater) Run(terminated chan bool, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)

	go func() {
		<-u.ctx.Done()
		ticker.Stop()
		log.Println("Metrics update stopped.")
		terminated <- true
	}()

	for range ticker.C {
		u.updateMetrics()
	}
}

func (u *updater) GetMetrics() ([]models.Metrics, error) {
	u.mx.Lock()
	defer u.mx.Unlock()
	return u.metrics.GetAll()
}

func (u *updater) updateMetrics() {
	runtime.ReadMemStats(u.memStats)

	u.mx.Lock()
	defer u.mx.Unlock()

	u.updateGauge("Alloc", float64(u.memStats.Alloc))
	u.updateGauge("BuckHashSys", float64(u.memStats.BuckHashSys))
	u.updateGauge("Frees", float64(u.memStats.Frees))
	u.updateGauge("GCCPUFraction", u.memStats.GCCPUFraction)
	u.updateGauge("GCSys", float64(u.memStats.GCSys))
	u.updateGauge("HeapAlloc", float64(u.memStats.HeapAlloc))
	u.updateGauge("HeapIdle", float64(u.memStats.HeapIdle))
	u.updateGauge("HeapInuse", float64(u.memStats.HeapInuse))
	u.updateGauge("HeapObjects", float64(u.memStats.HeapObjects))
	u.updateGauge("HeapReleased", float64(u.memStats.HeapReleased))
	u.updateGauge("HeapSys", float64(u.memStats.HeapSys))
	u.updateGauge("LastGC", float64(u.memStats.LastGC))
	u.updateGauge("Lookups", float64(u.memStats.Lookups))
	u.updateGauge("MCacheInuse", float64(u.memStats.MCacheInuse))
	u.updateGauge("MCacheSys", float64(u.memStats.MCacheSys))
	u.updateGauge("MSpanInuse", float64(u.memStats.MSpanInuse))
	u.updateGauge("MSpanSys", float64(u.memStats.MSpanSys))
	u.updateGauge("Mallocs", float64(u.memStats.Mallocs))
	u.updateGauge("NextGC", float64(u.memStats.NextGC))
	u.updateGauge("NumForcedGC", float64(u.memStats.NumForcedGC))
	u.updateGauge("NumGC", float64(u.memStats.NumGC))
	u.updateGauge("OtherSys", float64(u.memStats.OtherSys))
	u.updateGauge("PauseTotalNs", float64(u.memStats.PauseTotalNs))
	u.updateGauge("StackInuse", float64(u.memStats.StackInuse))
	u.updateGauge("StackSys", float64(u.memStats.StackSys))
	u.updateGauge("Sys", float64(u.memStats.Sys))
	u.updateGauge("TotalAlloc", float64(u.memStats.TotalAlloc))

	u.updateCounter("PollCount")

	u.updateGauge("RandomValue", rand.Float64()) // nolint: gosec

	log.Println("Metrics has been updated.")
}

func (u *updater) updateCounter(id string) {
	one := int64(1)
	m := models.Metrics{
		ID:    id,
		MType: models.MCounter,
		Delta: &one, // increment by 1 in `.Update`
	}

	updErr := u.metrics.Update(m)
	if updErr != nil {
		log.Printf("can't update %+v\n", m)
	}
}

func (u *updater) updateGauge(id string, val float64) {
	m := models.Metrics{
		ID:    id,
		MType: models.MGauge,
		Value: &val,
	}

	updErr := u.metrics.Update(m)
	if updErr != nil {
		log.Printf("can't update %+v\n", m)
	}
}
