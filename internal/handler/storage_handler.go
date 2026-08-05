package handler

import (
	"fmt"
	"net/http"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type StorageHandler struct {
	storage repository.StoragePinger
	log     *zap.Logger
}

func NewStorageHandler(storage repository.StoragePinger, log *zap.Logger) *StorageHandler {
	return &StorageHandler{
		storage: storage,
		log:     log,
	}
}

func (ph *StorageHandler) HandlePing(c echo.Context) error {
	err := ph.storage.Ping(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to ping Db: %v", err))
	}
	return c.String(http.StatusOK, "Ping success!")
}
