package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/agent"
	"github.com/KonstantinPavlov/metric-service/internal/config"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	err := parseFlags()
	if err != nil {
		panic(err)
	}
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	storage := repository.NewMemStorage()
	provider := service.NewDefaultProvider(storage, zapLogger)
	collector := agent.NewMetricCollector(provider, zapLogger)

	collector.Start(ctx, time.Duration(flagPollInterval)*time.Second)
	exporter, err := agent.NewMetricsExporter(flagServerAddr, provider, http.Client{}, zapLogger, flagCryptoKey, flagPublicCryptoKeyPath, flagRateLimit)
	if err != nil {
		zapLogger.Error("failed to create exporter!", zap.Error(err))
		return
	}
	err = exporter.Start(ctx, time.Duration(flagReportInterval)*time.Second)
	if err != nil {
		zapLogger.Fatal("failed to start exporter!", zap.Error(err))
		return
	}
	<-ctx.Done()
	zapLogger.Info("shutting down agent gracefully...")
	collector.Stop()

	// special shutdown ctx!
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := exporter.Stop(shutdownCtx); err != nil {
		zapLogger.Error("failed to flush metrics during shutdown", zap.Error(err))
	} else {
		zapLogger.Info("all metrics successfully exported")
	}
	
}
