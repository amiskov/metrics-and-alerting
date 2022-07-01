package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func main() {

	// 1. ReamMemStas
	// 2. Send everything is separate goroutines.
	// 3. Update counter, finish everything.
	// 4. Start again.

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	fmt.Println(
		m.Alloc,
		m.BuckHashSys,
		m.Frees,
		m.GCCPUFraction,
		m.GCSys,
		m.HeapAlloc,
		m.HeapIdle,
		m.HeapInuse,
		m.HeapObjects,
		m.HeapReleased,
		m.HeapSys,
		m.LastGC,
		m.Lookups,
		m.MCacheInuse,
		m.MCacheSys,
		m.MSpanInuse,
		m.MSpanSys,
		m.Mallocs,
		m.NextGC,
		m.NumForcedGC,
		m.NumGC,
		m.OtherSys,
		m.PauseTotalNs,
		m.StackInuse,
		m.StackSys,
		m.Sys,
		m.TotalAlloc,
	)
}
