package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

type MetricHandler struct {
	Repository repository.MetricRepository
}

type ListView struct {
	Name  string
	Type  string
	Value interface{}
}

func (h *MetricHandler) HandleList(c echo.Context) error {
	counters := h.Repository.GetCounters()
	gauges := h.Repository.GetGauges()

	metricsData := make([]ListView, 0)
	for key, value := range counters {
		metricsData = append(metricsData, ListView{
			Name:  key,
			Type:  MetricCounter,
			Value: value},
		)
	}
	for key, value := range gauges {
		metricsData = append(metricsData, ListView{
			Name:  key,
			Type:  MetricGauge,
			Value: value},
		)
	}

	data := map[string]interface{}{
		"Metrics": metricsData,
	}

	return c.Render(http.StatusOK, "list-view.html", data)
}

func (h *MetricHandler) HandleValue(c echo.Context) error {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch metricType {
	case MetricCounter:
		value, ok := h.Repository.GetCounters()[metricName]
		if !ok {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		return c.String(http.StatusOK, fmt.Sprintf("%v", value))
	case MetricGauge:
		value, ok := h.Repository.GetGauges()[metricName]
		if !ok {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		return c.String(http.StatusOK, fmt.Sprintf("%v", value))
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}
}

func (h *MetricHandler) HandleUpdate(c echo.Context) error {
	metricType := c.Param("type")
	metricName := c.Param("name")
	value := c.Param("value")

	if metricName == "" {
		return c.String(http.StatusNotFound, "Metric name must be set")
	}

	switch metricType {
	case MetricCounter:
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "value must be a number")
		}
		log.Default().Printf("Saving data for counter metric %v", metricName)
		err = h.Repository.SaveCounter(metricName, vInt)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	case MetricGauge:
		vFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Failed to parse gauge value: %v", err))
		}
		err = h.Repository.SaveGauge(metricName, vFloat)
		log.Default().Printf("Saving data for gauge metric %v", metricName)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	default:
		return c.String(http.StatusBadRequest, "unkwnown metric type!")
	}
	return c.String(http.StatusOK, "metric saved")
}
