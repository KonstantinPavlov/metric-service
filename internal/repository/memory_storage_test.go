package repository

import (
	"testing"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.NotNil(t, storage.Counters)
	assert.NotNil(t, storage.GetNames(model.Counter))
	assert.NotNil(t, storage.GetNames(model.Gauge))
	assert.NotNil(t, storage.Gauges)

}

func TestSaveCounter(t *testing.T) {
	storage := NewMemStorage()
	metricName := "poll_count"
	// Первый вызов — сохраняем 5
	err := storage.SaveCounter(metricName, 5)
	if err != nil {
		t.Errorf("Not expected error in SaveCounter: %v", err)
	}
	metric := storage.GetCounter(metricName)
	assert.NotNil(t, metric)
	assert.Equal(t, int64(5), metric.Value, "Expected value 5")

	_ = storage.SaveCounter(metricName, 10)
	metric = storage.GetCounter(metricName)
	assert.NotNil(t, metric)
	assert.Equal(t, int64(15), metric.Value, "Expected value 15")

	metric = storage.GetCounter("unknown_metric")
	assert.Nil(t, metric)
}

func TestSaveGauge(t *testing.T) {
	storage := NewMemStorage()
	metricName := "alloc_value"

	err := storage.SaveGauge(metricName, 123.45)
	if err != nil {
		t.Errorf("Not expected error in SaveGauge: %v", err)
	}
	metric := storage.GetGauge(metricName)
	assert.NotNil(t, metric)
	assert.Equal(t, 123.45, metric.Value, "Expected value 15")

	_ = storage.SaveGauge(metricName, 500.1)
	metric = storage.GetGauge(metricName)
	assert.NotNil(t, metric)
	assert.Equal(t, 500.1, metric.Value, "Expected value 15")

	metric = storage.GetGauge("unknown_metric")
	assert.Nil(t, metric)
}
