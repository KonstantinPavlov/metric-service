package agent

import (
	"context"
	"testing"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"github.com/KonstantinPavlov/metric-service/internal/testutils"

	"github.com/stretchr/testify/assert"
)


func TestMetricsCollector_Start(t *testing.T) {
	calledChan := make(chan struct{})
	mock := &testutils.MockProvider{CalledChan: calledChan}

	mc := MetricsCollector{
		Provider: mock,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mc.Start(ctx, 10*time.Millisecond)

	for i := 1; i <= 2; i++ {
		select {
		case <-calledChan:
			// do nothing - was called
		case <-time.After(50 * time.Millisecond):
			assert.Failf(t, "Timeout", "Metric collection not executed at tick №%d", i)
			return
		}
	}
}

func TestMetricCollector_Stop(t *testing.T) {

	calledChan := make(chan struct{})
	mock := &testutils.MockProvider{CalledChan: calledChan}
	mc := MetricsCollector{
		Provider: mock,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mc.Start(ctx, 10*time.Millisecond)

	cancel()

	assert.Eventually(t, func() bool {
		waitFinished := make(chan struct{})
		go func() {
			mc.wg.Wait()
			close(waitFinished)
		}()

		select {
		case <-waitFinished:
			return true // WaitGroup in MetricCollector unlocked
		case <-time.After(10 * time.Millisecond):
			return false // Still waiting
		}
	}, 100*time.Millisecond, 10*time.Millisecond, "Gorutine must stop after cancel()")
}

func TestMetricsCollector_Collect(t *testing.T) {
	storage := repository.NewMemStorage()

	provider := service.DefaultProvider{
		Repository: storage,
	}

	collector := MetricsCollector{
		Provider: &provider,
	}

	provider.Repository.SaveCounter("PollCount", 5)

	collector.Collect()
	assert.Equal(t, int64(6), storage.Counters["PollCount"])

	assert.Contains(t, storage.Gauges, "RandomValue")
	randomVal := storage.Gauges["RandomValue"]
	assert.GreaterOrEqual(t, randomVal, 0.0)
	assert.Less(t, randomVal, 1.0)

	metricsToCheck := []string{"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}

	for _, metricName := range metricsToCheck {
		_, exists := storage.Gauges[metricName]
		assert.True(t, exists, "Metric %s must exists in Storage", metricName)
	}

	assert.Len(t, storage.Gauges, 28, "Total lenght of Gauges must be 28")
}
