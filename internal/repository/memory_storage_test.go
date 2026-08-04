package repository

import (
	"testing"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.NotNil(t, storage.Counters)
	counters, _ := storage.GetNames(t.Context(), model.Counter)
	assert.NotNil(t, counters)
	gauges, _ := storage.GetNames(t.Context(), model.Counter)
	assert.NotNil(t, gauges)
	assert.NotNil(t, storage.Gauges)

}

func TestSaveCounter(t *testing.T) {
	storage := NewMemStorage()
	metricName := "poll_count"
	// Первый вызов — сохраняем 5
	err := storage.SaveCounter(t.Context(), metricName, 5)
	if err != nil {
		t.Errorf("Not expected error in SaveCounter: %v", err)
	}
	metric, err := storage.GetCounter(t.Context(), metricName)
	assert.Nil(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, int64(5), metric.Value, "Expected value 5")

	_ = storage.SaveCounter(t.Context(), metricName, 10)
	metric, err = storage.GetCounter(t.Context(), metricName)
	assert.Nil(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, int64(15), metric.Value, "Expected value 15")

	metric, err = storage.GetCounter(t.Context(), "unknown_metric")
	assert.Nil(t, err)
	assert.Nil(t, metric)
}

func TestSaveGauge(t *testing.T) {
	storage := NewMemStorage()
	metricName := "alloc_value"

	err := storage.SaveGauge(t.Context(), metricName, 123.45)
	if err != nil {
		t.Errorf("Not expected error in SaveGauge: %v", err)
	}
	metric, err := storage.GetGauge(t.Context(), metricName)
	assert.Nil(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, 123.45, metric.Value, "Expected value 15")

	_ = storage.SaveGauge(t.Context(), metricName, 500.1)
	metric, err = storage.GetGauge(t.Context(), metricName)
	assert.Nil(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, 500.1, metric.Value, "Expected value 15")

	metric, err = storage.GetGauge(t.Context(), "unknown_metric")
	assert.Nil(t, err)
	assert.Nil(t, metric)
}
