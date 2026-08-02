package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinPavlov/metric-service/internal/model"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MockMetricRepository struct {
	SaveCounterFunc func(name string, value int64) error
	SaveGaugeFunc   func(name string, value float64) error
}

func (m *MockMetricRepository) Start() error {
	return nil
}

func (m *MockMetricRepository) Ping() error {
	return nil
}

func (m *MockMetricRepository) Stop() {

}

func (m *MockMetricRepository) GetNames(metricTYpe string) ([]string, error) {
	return make([]string, 0), nil
}

func (m *MockMetricRepository) GetCounter(name string) (*repository.MetricData, error) {
	if name == "special-counter" {
		return &repository.MetricData{
			Name:  "special-counter",
			Value: *new(int64(42)),
		}, nil
	}

	return nil, nil
}
func (m *MockMetricRepository) GetGauge(name string) (*repository.MetricData, error) {
	if name == "special-gauge" {
		return &repository.MetricData{
			Name:  "special-gauge",
			Value: *new(float64(42.42)),
		}, nil
	}

	return nil, nil
}

func (m *MockMetricRepository) SaveCounter(name string, value int64) error {
	if m.SaveCounterFunc != nil {
		return m.SaveCounterFunc(name, value)
	}
	return nil
}

func (m *MockMetricRepository) SaveGauge(name string, value float64) error {
	if m.SaveGaugeFunc != nil {
		return m.SaveGaugeFunc(name, value)
	}
	return nil
}

func TestMetricHandler_HandleParamUpdate(t *testing.T) {

	type testCase struct {
		name           string
		method         string
		url            string
		setupMock      func(m *MockMetricRepository)
		expectedStatus int
	}

	tests := []testCase{
		{
			name:           "StatusMethodNotAllowed not a POST method",
			method:         http.MethodGet,
			url:            "/update/counter/testMetric/10",
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "StatusOK counter",
			method: http.MethodPost,
			url:    "/update/counter/testCounter/10",
			setupMock: func(m *MockMetricRepository) {
				m.SaveCounterFunc = func(name string, value int64) error {
					if name != "testCounter" || value != 10 {
						t.Errorf("неверные параметры в SaveCounter: %s, %d", name, value)
					}
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "StatusOK gauge",
			method: http.MethodPost,
			url:    "/update/gauge/testGauge/5.5",
			setupMock: func(m *MockMetricRepository) {
				m.SaveGaugeFunc = func(name string, value float64) error {
					if name != "testGauge" || value != 5.5 {
						t.Errorf("неверные параметры в SaveGauge: %s, %f", name, value)
					}
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "StatusBadRequest - incorrect value",
			method:         http.MethodPost,
			url:            "/update/counter/testCounter/abc",
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "StatusBadRequest - incoreccret metric type",
			method:         http.MethodPost,
			url:            "/update/unknown/testMetric/10",
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "StatusInternalServerError",
			method: http.MethodPost,
			url:    "/update/counter/testCounter/10",
			setupMock: func(m *MockMetricRepository) {
				m.SaveCounterFunc = func(name string, value int64) error {
					return errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &MockMetricRepository{}
			tt.setupMock(mockRepo)

			handler := &MetricHandler{Repository: mockRepo, log: zapLogger}

			httpServer := echo.New()
			httpServer.POST("/update/:type/:name/:value", handler.HandleParamUpdate)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()
			httpServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, but was %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMetricHandler_HandleBodyUpdate(t *testing.T) {

	type testCase struct {
		name           string
		method         string
		body           model.Metrics
		setupMock      func(m *MockMetricRepository)
		expectedStatus int
	}

	tests := []testCase{
		{
			name:   "StatusMethodNotAllowed not a POST method",
			method: http.MethodGet,
			body: model.Metrics{
				ID:    "testMetric",
				MType: "counter",
				Delta: new(int64(10)),
			},
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "StatusOK counter",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testCounter",
				MType: "counter",
				Delta: new(int64(10)),
			},
			setupMock: func(m *MockMetricRepository) {
				m.SaveCounterFunc = func(name string, value int64) error {
					if name != "testCounter" || value != 10 {
						t.Errorf("неверные параметры в SaveCounter: %s, %d", name, value)
					}
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "StatusOK gauge",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testGauge",
				MType: "gauge",
				Value: new(float64(5.5)),
			},
			setupMock: func(m *MockMetricRepository) {
				m.SaveGaugeFunc = func(name string, value float64) error {
					if name != "testGauge" || value != 5.5 {
						t.Errorf("неверные параметры в SaveGauge: %s, %f", name, value)
					}
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "StatusBadRequest - incorrect value",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testCounter",
				MType: "counter",
			},
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "StatusBadRequest - incoreccret metric type",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testMetric",
				MType: "unknown",
				Delta: new(int64(10)),
			},
			setupMock:      func(m *MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "StatusInternalServerError",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testCounter",
				MType: "counter",
				Delta: new(int64(10)),
			},
			setupMock: func(m *MockMetricRepository) {
				m.SaveCounterFunc = func(name string, value int64) error {
					return errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &MockMetricRepository{}
			tt.setupMock(mockRepo)

			handler := &MetricHandler{Repository: mockRepo, log: zapLogger}

			httpServer := echo.New()
			httpServer.POST("/update", handler.HandleBodyUpdate)
			jsonBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/update", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			httpServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, but was %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMetricHandler_HandlePostValue(t *testing.T) {

	type testCase struct {
		name           string
		method         string
		body           model.Metrics
		expectedStatus int
	}

	tests := []testCase{
		{
			name:   "StatusMethodNotAllowed not a POST method",
			method: http.MethodGet,
			body: model.Metrics{
				ID:    "testMetric",
				MType: "counter",
				Delta: new(int64(10)),
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "StatusNotFound counter",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testCounter",
				MType: "counter",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "StatusNotFound gauge",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testGauge",
				MType: "gauge",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "StatusOK counter",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "special-counter",
				MType: "counter",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "StatusNotFound gauge",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testGauge",
				MType: "gauge",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "StatusOK gauge",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "special-gauge",
				MType: "gauge",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "StatusBadRequest - incoreccret metric type",
			method: http.MethodPost,
			body: model.Metrics{
				ID:    "testMetric",
				MType: "unknown",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &MockMetricRepository{}

			handler := &MetricHandler{Repository: mockRepo, log: zapLogger}

			httpServer := echo.New()
			httpServer.POST("/value/", handler.HandlePostValue)
			jsonBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/value/", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			httpServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, but was %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
