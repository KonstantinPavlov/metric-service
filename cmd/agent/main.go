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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage := repository.NewMemStorage()
	provider := &service.DefaultProvider{
		Repository: storage,
	}
	collector := agent.MetricsCollector{
		Provider: provider,
	}
	collector.Start(ctx, 2*time.Second)
	exporter := agent.MetricsExporter{
		Provider: provider,
		Client:   http.Client{},
	}
	exporter.Start(ctx, 10*time.Second)

	<-ctx.Done()
	exporter.Stop()
	collector.Stop()
}
