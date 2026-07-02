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

func (mc *MetricsCollector) Start(ctx context.Context, interval time.Duration) {
	mc.wg.Add(1)

	go func() {
		defer mc.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Default().Print("Start collecting metrics...")
				mc.Collect()
				log.Default().Print("End collecting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (mc *MetricsCollector) Stop() {
	mc.wg.Wait()
}


func (mc *MetricsCollector) Collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	mc.Provider.SaveGauge("Alloc", float64(ms.Alloc))
	mc.Provider.SaveGauge("BuckHashSys", float64(ms.BuckHashSys))
	mc.Provider.SaveGauge("Frees", float64(ms.Frees))
	mc.Provider.SaveGauge("GCCPUFraction", ms.GCCPUFraction)
	mc.Provider.SaveGauge("GCSys", float64(ms.GCSys))
	mc.Provider.SaveGauge("HeapAlloc", float64(ms.HeapAlloc))
	mc.Provider.SaveGauge("HeapIdle", float64(ms.HeapIdle))
	mc.Provider.SaveGauge("HeapInuse", float64(ms.HeapInuse))
	mc.Provider.SaveGauge("HeapObjects", float64(ms.HeapObjects))
	mc.Provider.SaveGauge("HeapReleased", float64(ms.HeapReleased))
	mc.Provider.SaveGauge("HeapSys", float64(ms.HeapSys))
	mc.Provider.SaveGauge("LastGC", float64(ms.LastGC))
	mc.Provider.SaveGauge("Lookups", float64(ms.Lookups))
	mc.Provider.SaveGauge("MCacheInuse", float64(ms.MCacheInuse))
	mc.Provider.SaveGauge("MCacheSys", float64(ms.MCacheSys))
	mc.Provider.SaveGauge("MSpanInuse", float64(ms.MSpanInuse))
	mc.Provider.SaveGauge("MSpanSys", float64(ms.MSpanSys))
	mc.Provider.SaveGauge("Mallocs", float64(ms.Mallocs))
	mc.Provider.SaveGauge("NextGC", float64(ms.NextGC))
	mc.Provider.SaveGauge("NumForcedGC", float64(ms.NumForcedGC))
	mc.Provider.SaveGauge("NumGC", float64(ms.NumGC))
	mc.Provider.SaveGauge("OtherSys", float64(ms.OtherSys))
	mc.Provider.SaveGauge("PauseTotalNs", float64(ms.PauseTotalNs))
	mc.Provider.SaveGauge("StackInuse", float64(ms.StackInuse))
	mc.Provider.SaveGauge("StackSys", float64(ms.StackSys))
	mc.Provider.SaveGauge("Sys", float64(ms.Sys))
	mc.Provider.SaveGauge("TotalAlloc", float64(ms.TotalAlloc))
	mc.Provider.SaveGauge("RandomValue", rand.Float64())
	mc.Provider.SaveCounter("PollCount", 1)
}