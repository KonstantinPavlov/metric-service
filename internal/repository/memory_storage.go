package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/KonstantinPavlov/metric-service/internal/model"
)

type MemStorage struct {
	mu       sync.RWMutex
	Counters map[string]int64
	Gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}
}

func (ms *MemStorage) Ping(_ context.Context) error {
	return nil
}

func (ms *MemStorage) Start(_ context.Context) error {
	return nil
}

func (ms *MemStorage) Stop() {

}

func (ms *MemStorage) GetNames(_ context.Context, metricType string) ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	switch metricType {
	case model.Counter:
		return getMapKeys(ms.Counters), nil
	case model.Gauge:
		return getMapKeys(ms.Gauges), nil
	}
	return make([]string, 0), nil
}

func (ms *MemStorage) GetCounter(_ context.Context, name string) (*MetricData, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Counters[name]
	if !ok {
		return nil, nil
	}
	return &MetricData{
		Name:  name,
		Value: val,
	}, nil
}

func (ms *MemStorage) GetGauge(_ context.Context, name string) (*MetricData, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Gauges[name]
	if !ok {
		return nil, nil
	}
	return &MetricData{
		Name:  name,
		Value: val,
	}, nil
}

func (ms *MemStorage) SaveCounter(_ context.Context, name string, value int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Counters[name] += value
	return nil
}

func (ms *MemStorage) SaveGauge(_ context.Context, name string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Gauges[name] = value
	return nil
}

func (ms *MemStorage) SaveMetrics(
	ctx context.Context,
	counters []MetricData,
	gauges []MetricData,
) error {
	for _, counter := range counters {
		if metricValue, ok := counter.Value.(int64); ok {
			ms.SaveCounter(ctx, counter.Name, metricValue)
		} else {
			return fmt.Errorf("Value is not a int64!")
		}
	}
	for _, counter := range gauges {
		if metricValue, ok := counter.Value.(float64); ok {
			ms.SaveGauge(ctx, counter.Name, metricValue)
		} else {
			return fmt.Errorf("Value is not a float64!")
		}
	}
	return nil
}

func getMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
