package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
)

type MetricHandler struct {
	Repository repository.MetricRepository
}

type ListView struct {
	Name  string
	Type  string
	Value interface{}
}

func appenListView(views []ListView, metricType string, metric repository.MetricData) []ListView {
	return append(views, ListView{
		Name:  metric.Name,
		Type:  metricType,
		Value: metric.Value},
	)
}

func (h *MetricHandler) HandleList(c echo.Context) error {

	counterNames := h.Repository.GetNames(model.Counter)
	counters := make([]repository.MetricData, 0)

	for _, counter := range counterNames {
		metric := h.Repository.GetCounter(counter)
		if metric != nil {
			counters = append(counters, *metric)
		}
	}
	gaugesNames := h.Repository.GetNames(model.Gauge)
	gauges := make([]repository.MetricData, 0)

	for _, gauge := range gaugesNames {
		metric := h.Repository.GetGauge(gauge)
		if metric != nil {
			gauges = append(gauges, *metric)
		}
	}

	metricsData := make([]ListView, 0)
	for _, metric := range counters {
		metricsData = appenListView(metricsData,model.Counter, metric)
	}

	for _, metric := range gauges {
		metricsData = appenListView(metricsData,model.Gauge, metric)
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
	case model.Counter:
		metric := h.Repository.GetCounter(metricName)
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		return c.String(http.StatusOK, fmt.Sprintf("%v", metric.Value))
	case model.Gauge:
		metric := h.Repository.GetGauge(metricName)
		if metric == nil {
			return c.String(http.StatusNotFound, "metric not found!")
		}
		return c.String(http.StatusOK, fmt.Sprintf("%v", metric.Value))
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
	case model.Counter:
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "value must be a number")
		}
		log.Default().Printf("Saving data for counter metric %v", metricName)
		err = h.Repository.SaveCounter(metricName, vInt)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save counter: %v", err))
		}
	case model.Gauge:
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
