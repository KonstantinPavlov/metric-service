package service

import (
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

func (provider *DefaultProvider) GetCounters() map[string]int64 {
	return provider.Repository.GetCounters()
}

func (provider *DefaultProvider) GetGauges() map[string]float64 {
	return provider.Repository.GetGauges()
}

func (provider *DefaultProvider) SaveCounter(name string, value int64) error {
	return provider.Repository.SaveCounter(name, value)
}

func (provider *DefaultProvider) SaveGauge(name string, value float64) error {
	return provider.Repository.SaveGauge(name, value)
}
