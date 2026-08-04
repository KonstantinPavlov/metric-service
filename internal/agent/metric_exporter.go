package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"go.uber.org/zap"
)

type MetricsExporter struct {
	serverUrl string
	provider  service.MetricsProvider
	client    http.Client
	log       *zap.Logger
	wg        sync.WaitGroup
}

func NewMetricsExporter(serverUrl string, provider service.MetricsProvider, client http.Client, log *zap.Logger) MetricsExporter {
	return MetricsExporter{
		serverUrl: serverUrl,
		provider:  provider,
		client:    client,
		log:       log,
	}
}

func (me *MetricsExporter) Start(ctx context.Context, interval time.Duration) {
	me.wg.Add(1)
	go func() {
		defer me.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				me.log.Info("Start exporting metrics...")
				me.Export(ctx)
				me.log.Info("End exporting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (me *MetricsExporter) Stop() {
	me.wg.Wait()
}

func (me *MetricsExporter) Export(ctx context.Context) {
	requests := make([]model.Metrics, 0)
	for key, value := range me.provider.GetCounters(ctx) {
		requests = append(requests, model.Metrics{
			ID:    key,
			MType: model.Counter,
			Delta: &value,
		})
	}
	for key, value := range me.provider.GetGauges(ctx) {
		requests = append(requests, model.Metrics{
			ID:    key,
			MType: model.Gauge,
			Value: &value,
		})
	}
	if len(requests) == 0 {
		me.log.Warn("Metrics are empty!")
		return
	}
	me.postMetrics(ctx, requests)
}

func (me *MetricsExporter) postMetrics(ctx context.Context, requests []model.Metrics) {
	jsonBytes, err := json.Marshal(requests)
	if err != nil {
		me.log.Error("Failed to marshall request!", zap.Error(err))
		return
	}
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(jsonBytes); err != nil {
		me.log.Error("Failed to compress data!", zap.Error(err))
		return
	}
	if err := gzWriter.Close(); err != nil {
		me.log.Error("Failed to close compress data!", zap.Error(err))
		return
	}

	me.log.Debug("Compressed data", zap.Int("before", len(jsonBytes)), zap.Int("after", compressed.Len()))
	compressedBytes := compressed.Bytes()
	maxAttemps := 4
	retryDelay := 1 * time.Second

	var resp *http.Response

	for attempt := 1; attempt <= maxAttemps; attempt++ {
		if err := ctx.Err(); err != nil {
			me.log.Error("Context cancelled during HTTP retries", zap.Error(err))
			return
		}
		bodyReader := bytes.NewReader(compressedBytes)
		request, err := http.NewRequestWithContext(ctx, "POST", "http://"+me.serverUrl+"/updates/", bodyReader)
		if err != nil {
			me.log.Error("Failed to create request!", zap.Error(err))
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		resp, err = me.client.Do(request)
		if err == nil {
			// all fine!
			break
		}
		if attempt == maxAttemps {
			me.log.Error("Error publishing metrics after all retries", zap.Int("size", len(requests)), zap.Error(err))
			return
		}

		me.log.Warn("Network error occurred, retrying...", zap.Int("attempt", attempt), zap.Error(err))

		select {
		case <-time.After(retryDelay):
			retryDelay += 2 * time.Second
		case <-ctx.Done():
			me.log.Error("Context cancelled while waiting for next retry", zap.Error(ctx.Err()))
			return
		}
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp != nil && resp.Header.Get("Content-Type") != "application/json" {
		me.log.Error("Server response content-type is not valid!")
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			me.log.Error("Error reading response body", zap.Error(err))
			return
		}
		me.log.Info("Server response body", zap.String("body", string(bodyBytes)))
	}
}
