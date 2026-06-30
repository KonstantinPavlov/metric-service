package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
)

type MockMetricRepository struct {
	SaveCounterFunc func(name string, value int64) error
	SaveGaugeFunc   func(name string, value float64) error
}

func (m *MockMetricRepository) GetNames(metricTYpe string) []string {
	return make([]string, 0)
}

func (m *MockMetricRepository) GetCounter(name string) *repository.MetricData {
	return nil
}
func (m *MockMetricRepository) GetGauge(name string) *repository.MetricData {
	return nil
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

func TestMetricHandler_HandleUpdate(t *testing.T) {

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &MockMetricRepository{}
			tt.setupMock(mockRepo)

			handler := &MetricHandler{Repository: mockRepo}

			httpServer := echo.New()
			httpServer.POST("/update/:type/:name/:value", handler.HandleUpdate)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()
			httpServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, but was %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
