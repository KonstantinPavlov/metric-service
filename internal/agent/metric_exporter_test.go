package agent

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"github.com/KonstantinPavlov/metric-service/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMetricsExporter_Start(t *testing.T) {
	calledChan := make(chan struct{})
	mock := &testutils.MockProvider{CalledChan: calledChan}

	mc := MetricsExporter{
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

func TestMetricsExporter_Stop(t *testing.T) {

	calledChan := make(chan struct{})
	mock := &testutils.MockProvider{CalledChan: calledChan}
	mc := MetricsExporter{
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

type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMetricsExporter_Export(t *testing.T) {
	callCounter := 0
	storage := repository.NewMemStorage()
	provider := service.DefaultProvider{
		Repository: storage,
	}
	provider.SaveCounter("some-counter", 1)
	provider.SaveGauge("some-gauge", 1)
	exporter := &MetricsExporter{
		Provider: &provider,
		Client: http.Client{
			Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				callCounter++
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			}),
		},
	}
	exporter.Export(context.Background())
	assert.Equal(t, 2, callCounter, "Export Method calling http.Client.Post must be 2 times - for gauge and for counter")
}
