package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// mockRepository реализует интерфейс MetricRepository для демонстрации работы примеров.
type mockRepository struct {
	repository.MetricRepository
}

func (m *mockRepository) SaveCounter(ctx context.Context, name string, value int64) error {
	return nil
}

func (m *mockRepository) GetCounter(ctx context.Context, name string) (*repository.MetricData, error) {
	if name == "PollCount" {
		return &repository.MetricData{Name: "PollCount", Value: int64(5)}, nil
	}
	return nil, nil
}

func (m *mockRepository) SaveMetrics(ctx context.Context, counters, gauges []repository.MetricData) error {
	return nil
}

// ExampleMetricHandler_HandlePostValue показывает, как отправлять POST-запрос
// для получения текущего значения метрики из репозитория.
func ExampleMetricHandler_HandlePostValue() {
	e := echo.New()
	
	// Готовим JSON-запрос для получения метрики Counter
	reqBody, _ := json.Marshal(model.Metrics{
		ID:    "PollCount",
		MType: model.Counter,
	})
	
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Инициализируем хэндлер с фейковым репозиторием и Nop-логгером
	mh := handler.NewMetricHandler(&mockRepository{}, zap.NewNop())

	if err := mh.HandlePostValue(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Status:", rec.Code)
	fmt.Println("Response JSON:", strings.TrimSpace(rec.Body.String()))

	// Output:
	// Status: 200
	// Response JSON: {"id":"PollCount","type":"counter","delta":5}
}

// ExampleMetricHandler_HandleGetValue показывает, как извлечь значение
// метрики, переданное через URL-параметры.
func ExampleMetricHandler_HandleGetValue() {
	e := echo.New()
	
	// Симулируем GET-запрос к /value/counter/PollCount
	req := httptest.NewRequest(http.MethodGet, "/value/counter/PollCount", nil)
	rec := httptest.NewRecorder()
	
	c := e.NewContext(req, rec)
	c.SetPath("/value/:type/:name")
	c.SetParamNames("type", "name")
	c.SetParamValues("counter", "PollCount")

	mh := handler.NewMetricHandler(&mockRepository{}, zap.NewNop())

	if err := mh.HandleGetValue(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Status:", rec.Code)
	fmt.Println("Value:", rec.Body.String())

	// Output:
	// Status: 200
	// Value: 5
}

// ExampleMetricHandler_HandleBodyUpdate показывает процесс обновления
// одной конкретной метрики через JSON-тело запроса.
func ExampleMetricHandler_HandleBodyUpdate() {
	e := echo.New()
	
	var delta int64 = 10
	reqBody, _ := json.Marshal(model.Metrics{
		ID:    "PollCount",
		MType: model.Counter,
		Delta: &delta,
	})
	
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mh := handler.NewMetricHandler(&mockRepository{}, zap.NewNop())

	if err := mh.HandleBodyUpdate(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Status:", rec.Code)
	
	// Output:
	// Status: 200
}

// ExampleMetricHandler_HandleUpdates демонстрирует пакетное (batch) обновление 
// нескольких метрик одной транзакцией.
func ExampleMetricHandler_HandleUpdates() {
	e := echo.New()
	
	var delta int64 = 1
	var value = 42.15
	
	metricsBatch := []model.Metrics{
		{ID: "PollCount", MType: model.Counter, Delta: &delta},
		{ID: "Alloc", MType: model.Gauge, Value: &value},
	}
	
	reqBody, _ := json.Marshal(metricsBatch)
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mh := handler.NewMetricHandler(&mockRepository{}, zap.NewNop())

	if err := mh.HandleUpdates(c); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Status:", rec.Code)
	
	// Output:
	// Status: 200
}