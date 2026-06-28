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
	"github.com/labstack/echo/v4"
)

func main() {
	parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage := repository.NewMemStorage()
	provider := &service.DefaultProvider{
		Repository: storage,
	}
	collector := agent.MetricsCollector{
		Provider: provider,
	}

	collector.Start(ctx, time.Duration(flagPollInterval)*time.Second)
	exporter := agent.MetricsExporter{
		Provider: provider,
		Client:   http.Client{},
	}
	exporter.Start(ctx, time.Duration(flagReportInterval)*time.Second)
	httpServer := echo.New()
	go func() {
		if err := httpServer.Start(flagRunAddr); err != nil && err != http.ErrServerClosed {
			httpServer.Logger.Fatal("Failed to run server: ", err)
		}
	}()
	<-ctx.Done()
	exporter.Stop()
	collector.Stop()
	err := httpServer.Shutdown(ctx)
	if err != nil {
		httpServer.Logger.Fatal("Failed to shutdown server")

	}
}
