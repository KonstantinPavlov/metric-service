package agent

import (
	"context"
	"log"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/service"
)

type MetricsCollector struct {
	Provider service.MetricsProvider
	wg       sync.WaitGroup
}

func (me *MetricsCollector) Start(ctx context.Context, interval time.Duration) {
	me.wg.Add(1)

	go func() {
		defer me.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Default().Print("Start collecting metrics...")
				me.Collect()
				log.Default().Print("End collecting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (me *MetricsCollector) Stop() {
	me.wg.Wait()
}


func (me *MetricsCollector) Collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	me.Provider.SaveGauge("Alloc", float64(ms.Alloc))
	me.Provider.SaveGauge("BuckHashSys", float64(ms.BuckHashSys))
	me.Provider.SaveGauge("Frees", float64(ms.Frees))
	me.Provider.SaveGauge("GCCPUFraction", ms.GCCPUFraction)
	me.Provider.SaveGauge("GCSys", float64(ms.GCSys))
	me.Provider.SaveGauge("HeapAlloc", float64(ms.HeapAlloc))
	me.Provider.SaveGauge("HeapIdle", float64(ms.HeapIdle))
	me.Provider.SaveGauge("HeapInuse", float64(ms.HeapInuse))
	me.Provider.SaveGauge("HeapObjects", float64(ms.HeapObjects))
	me.Provider.SaveGauge("HeapReleased", float64(ms.HeapReleased))
	me.Provider.SaveGauge("HeapSys", float64(ms.HeapSys))
	me.Provider.SaveGauge("LastGC", float64(ms.LastGC))
	me.Provider.SaveGauge("Lookups", float64(ms.Lookups))
	me.Provider.SaveGauge("MCacheInuse", float64(ms.MCacheInuse))
	me.Provider.SaveGauge("MCacheSys", float64(ms.MCacheSys))
	me.Provider.SaveGauge("MSpanInuse", float64(ms.MSpanInuse))
	me.Provider.SaveGauge("MSpanSys", float64(ms.MSpanSys))
	me.Provider.SaveGauge("Mallocs", float64(ms.Mallocs))
	me.Provider.SaveGauge("NextGC", float64(ms.NextGC))
	me.Provider.SaveGauge("NumForcedGC", float64(ms.NumForcedGC))
	me.Provider.SaveGauge("NumGC", float64(ms.NumGC))
	me.Provider.SaveGauge("OtherSys", float64(ms.OtherSys))
	me.Provider.SaveGauge("PauseTotalNs", float64(ms.PauseTotalNs))
	me.Provider.SaveGauge("StackInuse", float64(ms.StackInuse))
	me.Provider.SaveGauge("StackSys", float64(ms.StackSys))
	me.Provider.SaveGauge("Sys", float64(ms.Sys))
	me.Provider.SaveGauge("TotalAlloc", float64(ms.TotalAlloc))
	me.Provider.SaveGauge("RandomValue", rand.Float64())
	me.Provider.SaveCounter("PollCount", 1)
}