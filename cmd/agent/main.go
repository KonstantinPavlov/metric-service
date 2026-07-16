package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/agent"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	err := parseFlags()
	if err != nil {
		panic(err)
	}
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage := repository.NewMemStorage()
	provider := &service.DefaultProvider{
		Repository: storage,
	}
	collector := agent.NewMetricCollector(provider, zapLogger)

	collector.Start(ctx, time.Duration(flagPollInterval)*time.Second)
	exporter := agent.NewMetricsExporter(flagServerAddr, provider, http.Client{}, zapLogger)
	exporter.Start(ctx, time.Duration(flagReportInterval)*time.Second)
	<-ctx.Done()
	exporter.Stop()
	collector.Stop()
}
