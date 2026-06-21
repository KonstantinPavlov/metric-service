package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

type MetricHandler struct {
	Repository repository.MetricRepository
}

func (h *MetricHandler) HandleUpdate(w http.ResponseWriter, req *http.Request) {
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
		log.Default().Printf("Saving data for counter metric %v", metricName)
		err = h.Repository.SaveCounter(metricName, vInt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case MetricGauge:
		vFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = h.Repository.SaveGauge(metricName, vFloat)
		log.Default().Printf("Saving data for gauge metric %v", metricName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
