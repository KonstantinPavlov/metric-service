package repository

import "context"

type MetricRepository interface {
	GetNames(ctx context.Context, metricType string) ([]string, error)
	GetCounter(ctx context.Context, name string) (*MetricData, error)
	GetGauge(ctx context.Context, name string) (*MetricData, error)
	SaveCounter(ctx context.Context, name string, value int64) error
	SaveGauge(ctx context.Context, name string, value float64) error
	SaveMetrics(
		ctx context.Context,
		counters []MetricData,
		gauges []MetricData,
	) error
	Ping(ctx context.Context) error
	Start(ctx context.Context) error
	Stop()	
}

type MetricData struct {
	Name  string
	Value interface{}
}
