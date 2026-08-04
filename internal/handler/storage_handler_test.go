package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func TestHandlePing(t *testing.T) {
	logger := zap.NewNop()
	t.Run("success ping", func(t *testing.T) {
		mockStorage := new(MockStorage)
		mockStorage.On("Ping").Return(nil)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := NewStorageHandler(mockStorage, logger)
		err := h.HandlePing(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Ping success!", rec.Body.String())
		mockStorage.AssertExpectations(t)
	})

	t.Run("failed ping", func(t *testing.T) {
		mockStorage := new(MockStorage)
		dbErr := errors.New("connection refused")
		mockStorage.On("Ping").Return(dbErr)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := NewStorageHandler(mockStorage, logger)
		err := h.HandlePing(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to ping Db: connection refused")
		mockStorage.AssertExpectations(t)
	})
}
