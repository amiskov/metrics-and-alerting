// Package `updater` creates a snapshot of a running service with a bunch of metrics.
package updater

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
	"github.com/amiskov/metrics-and-alerting/pkg/storage/inmem"
)

type store interface {
	Update(m models.Metrics) error
}

type updater struct {
	ctx          context.Context
	terminated   chan<- bool
	memStats     *runtime.MemStats
	metrics      store
	pollInterval time.Duration
}

func New(ctx context.Context, terminated chan<- bool, db *inmem.DB, pollInterval time.Duration) *updater {
	return &updater{
		ctx:          ctx,
		terminated:   terminated,
		memStats:     new(runtime.MemStats),
		metrics:      db,
		pollInterval: pollInterval,
	}
}

// Run the process which takes the metrics snapshot once in interval.
func (u *updater) Run() {
	ticker := time.NewTicker(u.pollInterval)

	wg := new(sync.WaitGroup)

	go func() {
		<-u.ctx.Done()
		ticker.Stop()
		log.Println("Metrics updater stopped.")
		u.terminated <- true
	}()

	for range ticker.C {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u.updateMemStats()
			logger.Log(u.ctx).Info("Mem stats was updated.")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			u.updateVirtualMem()
			logger.Log(u.ctx).Info("Virtual mem was updated.")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			u.updateCPU()
			logger.Log(u.ctx).Info("CPU was updated.")
		}()

		wg.Wait()
		log.Println("Metrics updated.")
	}
}

func (u *updater) updateMemStats() {
	runtime.ReadMemStats(u.memStats)

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
}

func (u *updater) updateVirtualMem() {
	vMem, err := mem.VirtualMemory()
	if err != nil {
		logger.Log(u.ctx).Error("can't create virtual memory stats object: %v", err)
		return
	}
	u.updateGauge("TotalMemory", float64(vMem.Total))
	u.updateGauge("FreeMemory", float64(vMem.Available))
}

func (u *updater) updateCPU() {
	cpu, err := cpu.Percent(0, true)
	if err != nil {
		logger.Log(u.ctx).Error("failed getting CPU info: %v", err)
		return
	}
	u.updateGauge("CPUutilization1", cpu[1])
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
