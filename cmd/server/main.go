package main

import (
	"net/http"
	"strconv"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, storage.handleUpdate)

	return http.ListenAndServe(":8080", mux)
}

type MemStorage struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[string]int64),
		make(map[string]float64),
	}
}

func (storage *MemStorage) handleUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	metricType := req.PathValue("type")
	metricName := req.PathValue("name")
	value := req.PathValue("value")

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case MetricCounter:
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.saveCounter(vInt, metricName)
	case MetricGauge:
		vFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.saveGauge(vFloat, metricName)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (storage *MemStorage) saveCounter(value int64, metric string) {
	storage.Counters[metric] += value
}

func (storage *MemStorage) saveGauge(value float64, metric string) {
	storage.Gauges[metric] = value
}
