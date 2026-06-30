package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/model"
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
				me.Export(ctx)
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

func (me *MetricsExporter) Export(ctx context.Context) {
	for key, value := range me.Provider.GetCounters() {
		me.postMetric(ctx, model.Counter, key, fmt.Sprint(value))

	}
	for key, value := range me.Provider.GetGauges() {
		me.postMetric(ctx, model.Gauge, key, fmt.Sprint(value))
	}
}

func (me *MetricsExporter) postMetric(ctx context.Context, metricType string, name string, value string) {

	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://"+me.ServerUrl+"/update/%v/%v/%v", metricType, name, value), nil)
	request.Header.Set("Content-Type", "text/plain")
	if err != nil {
		log.Default().Printf("Failed to create request: %v", err)
	}

	_, err = me.Client.Do(request)

	if err != nil {
		log.Default().Printf("Error in publishing metric %v with type %v. Error: %v", name, metricType, err)
	}
}
