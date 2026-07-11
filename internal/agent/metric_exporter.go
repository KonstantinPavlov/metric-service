package agent

import (
	"context"
	"fmt"
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
		me.postMetric(ctx, model.Counter, key, fmt.Sprint(value))

	}
	for key, value := range me.provider.GetGauges() {
		me.postMetric(ctx, model.Gauge, key, fmt.Sprint(value))
	}
}

func (me *MetricsExporter) postMetric(ctx context.Context, metricType string, name string, value string) {

	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://"+me.serverUrl+"/update/%v/%v/%v", metricType, name, value), nil)
	request.Header.Set("Content-Type", "text/plain")
	if err != nil {
		me.log.Error("Failed to create request!", zap.Error(err))
	}

	_, err = me.client.Do(request)

	if err != nil {
		me.log.Error("Error publishing metric", zap.String("name", name), zap.String("metric_type", metricType), zap.Error(err))
	}
}
