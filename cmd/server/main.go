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

	_ "net/http/pprof"

	"github.com/KonstantinPavlov/metric-service/internal/audit"
	"github.com/KonstantinPavlov/metric-service/internal/config"
	"github.com/KonstantinPavlov/metric-service/internal/crypto"
	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/logger"
	"github.com/KonstantinPavlov/metric-service/internal/middleware"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

//go:embed views/*
var viewsFS embed.FS

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config.PrintBuildInfo(buildVersion, buildDate, buildCommit)
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

func defineStorage(zapLogger *zap.Logger) repository.MetricRepository {
	if cfg.DBDSN != "" {
		return repository.NewPgStorage(cfg.DBDSN, zapLogger)
	}
	memStorage := repository.NewMemStorage()
	if cfg.StorePath != "" {
		return repository.NewFileStorage(
			cfg.StoreIntervalSeconds,
			cfg.StorePath,
			cfg.Restore,
			memStorage,
			zapLogger,
		)
	}
	return memStorage
}

func run(zapLogger *zap.Logger) error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	storage := defineStorage(zapLogger)
	var storageErr error
	go func() {
		if err := storage.Start(ctx); err != nil {
			zapLogger.Error("Storage failed to start", zap.Error(err))
			storageErr = err
			stop()
		}
	}()

	webHandler := handler.NewMetricHandler(
		storage,
		zapLogger,
		1,
	)
	webHandler.Start(ctx)
	tmpl, err := template.ParseFS(viewsFS, "views/*.html")
	if err != nil {
		return err
	}

	renderer := &handler.TemplateRenderer{
		Template: tmpl,
	}

	storageHandler := handler.NewStorageHandler(storage, zapLogger)

	httpServer := echo.New()
	httpServer.Use(logger.ZapMiddleware(zapLogger))

	if cfg.CryptoKey != "" {
		keyBytes, err := crypto.DecodeKeyString(cfg.CryptoKey)
		if err != nil {
			return err
		}
		zapLogger.Info("Adding crypto middlewares...")
		httpServer.Use(middleware.Sha256RequestMiddleware(keyBytes))
		httpServer.Use(middleware.Sha256ResponseMiddleware(keyBytes))
	}

	if cfg.CryptoKeyPath != "" {
		middle, err := middleware.NewPrivateKeyMiddleware(cfg.CryptoKeyPath)
		if err != nil {
			return err
		}
		httpServer.Use(middle.Middleware())
	}

	// observers
	if cfg.AuditFile != "" {
		o, err := audit.NewAuditFileObserver(zapLogger, cfg.AuditFile)
		if err != nil {
			return err
		}
		webHandler.Register(o)
	}

	if cfg.AuditURL != "" {
		o := audit.NewRemoteAuditObserver(zapLogger, cfg.AuditURL)
		webHandler.Register(o)
	}

	httpServer.Use(echoMiddleware.Decompress())
	httpServer.Use(middleware.GzipMiddleware())
	httpServer.Renderer = renderer
	httpServer.POST("/update/:type/:name/:value", webHandler.HandleParamUpdate)
	httpServer.POST("/update/", webHandler.HandleBodyUpdate)
	httpServer.POST("/updates/", webHandler.HandleUpdates)
	httpServer.GET("/value/:type/:name", webHandler.HandleGetValue)
	httpServer.POST("/value/", webHandler.HandlePostValue)
	httpServer.GET("/", webHandler.HandleList)
	httpServer.GET("/ping", storageHandler.HandlePing)

	// pprof middleware
	pprof.Register(httpServer)

	go func() {
		zapLogger.Info("Starting Web server...")
		err := httpServer.Start(cfg.RunAddr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zapLogger.Fatal("Failed to start echo web server", zap.Error(err))
		}
	}()

	// gracefull shutdown
	<-ctx.Done()
	zapLogger.Info("Shutting down server gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// First is echo web server!
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("Web server shutdown failed", zap.Error(err))
	} else {
		zapLogger.Info("Web server stopped successfully (no active connections)")
	}
	// Second is web Handler
	webHandler.Stop()
	// And storage will be last - no more requests
	storage.Stop()

	if storageErr != nil {
		return storageErr
	}
	return nil
}
