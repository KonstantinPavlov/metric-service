package agent

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/service"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type MetricsCollector struct {
	Provider service.MetricsProvider
	log      *zap.Logger
	wg       sync.WaitGroup
}

func NewMetricCollector(provider service.MetricsProvider, log *zap.Logger) MetricsCollector {
	return MetricsCollector{
		Provider: provider,
		log:      log,
	}
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
				mc.log.Info("Start collecting metrics...")
				mc.Collect(ctx)
				mc.log.Info("End collecting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mc.log.Info("Start collecting gopsutil metrics...")
				mc.CollectPsUtil(ctx)
				mc.log.Info("End collecting gopsutil metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()

}

func (mc *MetricsCollector) Stop() {
	mc.wg.Wait()
}

func (mc *MetricsCollector) Collect(ctx context.Context) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	mc.Provider.SaveGauge(ctx, "Alloc", float64(ms.Alloc))
	mc.Provider.SaveGauge(ctx, "BuckHashSys", float64(ms.BuckHashSys))
	mc.Provider.SaveGauge(ctx, "Frees", float64(ms.Frees))
	mc.Provider.SaveGauge(ctx, "GCCPUFraction", ms.GCCPUFraction)
	mc.Provider.SaveGauge(ctx, "GCSys", float64(ms.GCSys))
	mc.Provider.SaveGauge(ctx, "HeapAlloc", float64(ms.HeapAlloc))
	mc.Provider.SaveGauge(ctx, "HeapIdle", float64(ms.HeapIdle))
	mc.Provider.SaveGauge(ctx, "HeapInuse", float64(ms.HeapInuse))
	mc.Provider.SaveGauge(ctx, "HeapObjects", float64(ms.HeapObjects))
	mc.Provider.SaveGauge(ctx, "HeapReleased", float64(ms.HeapReleased))
	mc.Provider.SaveGauge(ctx, "HeapSys", float64(ms.HeapSys))
	mc.Provider.SaveGauge(ctx, "LastGC", float64(ms.LastGC))
	mc.Provider.SaveGauge(ctx, "Lookups", float64(ms.Lookups))
	mc.Provider.SaveGauge(ctx, "MCacheInuse", float64(ms.MCacheInuse))
	mc.Provider.SaveGauge(ctx, "MCacheSys", float64(ms.MCacheSys))
	mc.Provider.SaveGauge(ctx, "MSpanInuse", float64(ms.MSpanInuse))
	mc.Provider.SaveGauge(ctx, "MSpanSys", float64(ms.MSpanSys))
	mc.Provider.SaveGauge(ctx, "Mallocs", float64(ms.Mallocs))
	mc.Provider.SaveGauge(ctx, "NextGC", float64(ms.NextGC))
	mc.Provider.SaveGauge(ctx, "NumForcedGC", float64(ms.NumForcedGC))
	mc.Provider.SaveGauge(ctx, "NumGC", float64(ms.NumGC))
	mc.Provider.SaveGauge(ctx, "OtherSys", float64(ms.OtherSys))
	mc.Provider.SaveGauge(ctx, "PauseTotalNs", float64(ms.PauseTotalNs))
	mc.Provider.SaveGauge(ctx, "StackInuse", float64(ms.StackInuse))
	mc.Provider.SaveGauge(ctx, "StackSys", float64(ms.StackSys))
	mc.Provider.SaveGauge(ctx, "Sys", float64(ms.Sys))
	mc.Provider.SaveGauge(ctx, "TotalAlloc", float64(ms.TotalAlloc))
	mc.Provider.SaveGauge(ctx, "RandomValue", rand.Float64())
	mc.Provider.SaveCounter(ctx, "PollCount", 1)
}

func (mc *MetricsCollector) CollectPsUtil(ctx context.Context) {
	vMemory, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		mc.Provider.SaveGauge(ctx, "TotalMemory", float64(vMemory.Total))
		mc.Provider.SaveGauge(ctx, "FreeMemory", float64(vMemory.Free))
	}
	cpuPercents, err := cpu.PercentWithContext(ctx, 100*time.Millisecond, true)
	if err == nil {
		for i, percent := range cpuPercents {
			metricName := "CPUutilization" + strconv.Itoa(i+1)
			mc.Provider.SaveGauge(ctx, metricName, float64(percent))
		}
	}
}
