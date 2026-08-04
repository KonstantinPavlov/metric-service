package service

import (
	"context"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

type MetricsProvider interface {
	GetCounters(ctx context.Context) map[string]int64
	GetGauges(ctx context.Context) map[string]float64
	SaveCounter(ctx context.Context, name string, value int64) error
	SaveGauge(ctx context.Context, name string, value float64) error
}

type DefaultProvider struct {
	repository repository.MetricRepository
	log        *zap.Logger
}

func NewDefaultProvider(repository repository.MetricRepository, log *zap.Logger) *DefaultProvider {
	return &DefaultProvider{
		repository: repository,
		log:        log,
	}
}

func (p *DefaultProvider) GetCounters(ctx context.Context) map[string]int64 {
	res := make(map[string]int64)
	counters, err := p.repository.GetNames(ctx, model.Counter)
	if err != nil {
		log.Error("Failed to getNames for Counters!", zap.Error(err))
		return res
	}
	for _, counter := range counters {
		metric, err := p.repository.GetCounter(ctx, counter)
		if err != nil {
			log.Error("Failed to get counter", zap.String("name", counter), zap.Error(err))
			continue
		}
		if metric != nil {
			val, ok := metric.Value.(int64)
			if ok {
				res[counter] = val
			}
		}

	}
	return res
}

func (p *DefaultProvider) GetGauges(ctx context.Context) map[string]float64 {
	res := make(map[string]float64)
	gauges, err := p.repository.GetNames(ctx, model.Gauge)
	if err != nil {
		log.Error("Failed to getNames for Gauges!", zap.Error(err))
		return res
	}
	for _, gauge := range gauges {
		metric, err := p.repository.GetGauge(ctx, gauge)
		if err != nil {
			log.Error("Failed to get gauge", zap.String("name", gauge), zap.Error(err))
			continue
		}
		if metric != nil {
			val, ok := metric.Value.(float64)
			if ok {
				res[gauge] = val
			}
		}
	}
	return res
}

func (p *DefaultProvider) SaveCounter(ctx context.Context, name string, value int64) error {
	return p.repository.SaveCounter(ctx, name, value)
}

func (p *DefaultProvider) SaveGauge(ctx context.Context, name string, value float64) error {
	return p.repository.SaveGauge(ctx, name, value)
}
