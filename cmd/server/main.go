package main

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/logger"
	"github.com/KonstantinPavlov/metric-service/internal/middleware"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	err := parseFlags()
	if err != nil {
		zapLogger.Fatal("Failed to parse configuration", zap.Error(err))
	}
	if err := run(zapLogger); err != nil {
		zapLogger.Fatal("Failed to start app", zap.Error(err))
	}
}

func defineStorage(ctx context.Context, zapLogger *zap.Logger) repository.MetricRepository {
	if flagDbDSN != "" {
		return repository.NewPgStorage(ctx, flagDbDSN, zapLogger)
	}
	memStorage := repository.NewMemStorage()
	if flagStorePath != "" {
		return repository.NewFileStorage(
			ctx,
			flagStoreIntervalSeconds,
			flagStorePath,
			flagRestore,
			memStorage,
			zapLogger,
		)
	}
	return memStorage
}

func run(zapLogger *zap.Logger) error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage := defineStorage(ctx, zapLogger)
	err := storage.Start()
	if err != nil {
		return err
	}
	// setup of echo web server
	webHandler := handler.NewMetricHandler(
		storage,
		zapLogger,
	)

	tmpl, err := template.ParseFS(viewsFS, "views/*.html")
	if err != nil {
		return err
	}

	renderer := &handler.TemplateRenderer{
		Template: tmpl,
	}

	pgHandler := handler.NewPgHandler(storage, zapLogger)

	httpServer := echo.New()
	httpServer.Use(logger.ZapMiddleware(zapLogger))
	httpServer.Use(echoMiddleware.Decompress())
	httpServer.Use(middleware.GzipMiddleware())
	httpServer.Renderer = renderer
	httpServer.POST("/update/:type/:name/:value", webHandler.HandleParamUpdate)
	httpServer.POST("/update/", webHandler.HandleBodyUpdate)
	httpServer.POST("/updates/", webHandler.HandleUpdates)
	httpServer.GET("/value/:type/:name", webHandler.HandleGetValue)
	httpServer.POST("/value/", webHandler.HandlePostValue)
	httpServer.GET("/", webHandler.HandleList)
	httpServer.GET("/ping", pgHandler.HandlePing)

	go func() {
		err := httpServer.Start(flagRunAddr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zapLogger.Fatal("Failed to start echo web server", zap.Error(err))
		}
	}()

	// gracefull shutdown
	<-ctx.Done()
	storage.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("Gracefull shutdown failed", zap.Error(err))
	}
	zapLogger.Info("Web server stopped")
	return nil
}
