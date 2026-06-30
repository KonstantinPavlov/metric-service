package service

import (
	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
)

type MetricsProvider interface {
	GetCounters() map[string]int64
	GetGauges() map[string]float64
	SaveCounter(name string, value int64) error
	SaveGauge(name string, value float64) error
}

type DefaultProvider struct {
	Repository repository.MetricRepository
}

func (p *DefaultProvider) GetCounters() map[string]int64 {
	res := make(map[string]int64)
	for _, counter := range p.Repository.GetNames(model.Counter) {
		metric := p.Repository.GetCounter(counter)
		if metric != nil {
			val, ok := metric.Value.(int64)
			if ok {
				res[counter] = val
			}
		}

	}
	return res
}

func (p *DefaultProvider) GetGauges() map[string]float64 {
	res := make(map[string]float64)
	for _, gauge := range p.Repository.GetNames(model.Gauge) {
		metric := p.Repository.GetGauge(gauge)
		if metric != nil {
			val, ok := metric.Value.(float64)
			if ok {
				res[gauge] = val
			}
		}
	}
	return res
}

func (p *DefaultProvider) SaveCounter(name string, value int64) error {
	return p.Repository.SaveCounter(name, value)
}

func (p *DefaultProvider) SaveGauge(name string, value float64) error {
	return p.Repository.SaveGauge(name, value)
}
