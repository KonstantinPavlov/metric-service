package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.NotNil(t, storage.Counters)
	assert.NotNil(t, storage.GetCounters())
	assert.NotNil(t, storage.Gauges)
	assert.NotNil(t, storage.GetGauges())
}

func TestSaveCounter(t *testing.T) {
	storage := NewMemStorage()
	metricName := "poll_count"
	// Первый вызов — сохраняем 5
	err := storage.SaveCounter(metricName, 5)
	if err != nil {
		t.Errorf("Not expected error in SaveCounter: %v", err)
	}
	val, exists := storage.GetCounters()[metricName]
	assert.True(t, exists, "Metric not found in storage!")
	assert.Equal(t, int64(5), val, "Expected value 5")

	_ = storage.SaveCounter(metricName, 10)
	val, exists = storage.GetCounters()[metricName]
	assert.True(t, exists, "Metric not found in storage!")
	assert.Equal(t, int64(15), val, "Expected value 15")
}

func TestSaveGauge(t *testing.T) {
	storage := NewMemStorage()
	metricName := "alloc_value"

	err := storage.SaveGauge(metricName, 123.45)
	if err != nil {
		t.Errorf("Not expected error in SaveGauge: %v", err)
	}
	val, exists := storage.GetGauges()[metricName]
	_ = storage.SaveGauge(metricName, 500.1)
	assert.True(t, exists, "Metric not found in storage!")
	assert.Equal(t, 123.45, val, "Expected value 15")

	val, exists = storage.GetGauges()[metricName]

	assert.True(t, exists, "Metric not found in storage!")
	assert.Equal(t, 500.1, val, "Expected value 15")
}
