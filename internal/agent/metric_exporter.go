package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/service"
)

type MetricsExporter struct {
	ServerUrl string
	Provider  service.MetricsProvider
	Client    http.Client
	wg        sync.WaitGroup
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
				log.Default().Print("Start exporting metrics...")
				me.Export()
				log.Default().Print("End exporting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (me *MetricsExporter) Stop() {
	me.wg.Wait()
}

func (exporter *MetricsExporter) Export() {
	for key, value := range exporter.Provider.GetCounters() {
		exporter.postMetric(handler.MetricCounter, key, fmt.Sprint(value))

	}
	for key, value := range exporter.Provider.GetGauges() {
		exporter.postMetric(handler.MetricGauge, key, fmt.Sprint(value))
	}
}

func (exporter *MetricsExporter) postMetric(metricType string, name string, value string) {
	_, err := exporter.Client.Post(fmt.Sprintf("http://"+exporter.ServerUrl+"/update/%v/%v/%v", metricType, name, value), "text/plain", nil)
	if err != nil {
		log.Default().Printf("Error in publishing metric %v with type %v. Error: %v", name, metricType, err)
	}
}
