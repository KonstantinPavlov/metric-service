package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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
	for key, value := range me.provider.GetCounters() {
		req := model.Metrics{
			ID:    key,
			MType: model.Counter,
			Delta: &value,
		}
		me.postMetric(ctx, req)

	}
	for key, value := range me.provider.GetGauges() {
		req := model.Metrics{
			ID:    key,
			MType: model.Gauge,
			Value: &value,
		}
		me.postMetric(ctx, req)
	}
}

func (me *MetricsExporter) postMetric(ctx context.Context, req model.Metrics) {
	jsonBytes, err := json.Marshal(req)
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
	request, err := http.NewRequestWithContext(ctx, "POST", "http://"+me.serverUrl+"/update/", &compressed)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	if err != nil {
		me.log.Error("Failed to create request!", zap.Error(err))
		return
	}

	resp, err := me.client.Do(request)

	if err != nil {
		me.log.Error("Error publishing metric", zap.String("name", req.ID), zap.String("metric_type", req.MType), zap.Error(err))
	}

	if resp != nil && resp.Header.Get("Content-Type") != "application/json" {
		me.log.Error("Server response content-type is not valid!")
	}
}
