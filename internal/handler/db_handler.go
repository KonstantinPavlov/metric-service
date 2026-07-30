package handler

import (
	"fmt"
	"net/http"

	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PgHandler struct {
	storage repository.StoragePinger
	log     *zap.Logger
}

func NewPgHandler(storage repository.StoragePinger, log *zap.Logger) *PgHandler {
	return &PgHandler{
		storage: storage,
		log:     log,
	}
}

func (ph *PgHandler) HandlePing(c echo.Context) error {
	err := ph.storage.Ping()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to ping Db: %v", err))
	}
	return c.String(http.StatusOK, "Ping success!")
}
