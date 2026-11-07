package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/audit"
	pb "github.com/KonstantinPavlov/metric-service/internal/proto"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockAuditObserver struct {
	events chan audit.AuditEvent
}

func (m *mockAuditObserver) Notify(_ context.Context, e audit.AuditEvent) {
	m.events <- e
}

func (m *mockAuditObserver) Stop() {}

func TestGRPCMetricsServer_UpdateMetrics(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	storage := repository.NewMemStorage()
	metricHandler := NewMetricHandler(storage, zapLogger, 1)

	ctx, cancel := context.WithCancel(context.Background())
	metricHandler.Start(ctx)
	// Важно: сначала cancel(), потом Stop(), чтобы воркер вышел из select и освободил WaitGroup
	defer func() {
		cancel()
		metricHandler.Stop()
	}()

	observer := &mockAuditObserver{events: make(chan audit.AuditEvent, 5)}
	metricHandler.Register(observer)

	grpcSrv := NewGRPCMetricsServer(&metricHandler)

	t.Run("success batch update", func(t *testing.T) {
		req := &pb.UpdateMetricsRequest{
			Metrics: []*pb.Metric{
				{
					Id:    "PollCount",
					Type:  pb.Metric_COUNTER,
					Delta: 10,
				},
				{
					Id:    "Alloc",
					Type:  pb.Metric_GAUGE,
					Value: 123.45,
				},
			},
		}

		md := metadata.Pairs("x-real-ip", "192.168.1.50")
		reqCtx := metadata.NewIncomingContext(ctx, md)

		resp, err := grpcSrv.UpdateMetrics(reqCtx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Проверяем, что данные сохранились в репозиторий
		counter, err := storage.GetCounter(ctx, "PollCount")
		assert.NoError(t, err)
		assert.NotNil(t, counter)
		assert.Equal(t, int64(10), counter.Value)

		gauge, err := storage.GetGauge(ctx, "Alloc")
		assert.NoError(t, err)
		assert.NotNil(t, gauge)
		assert.Equal(t, 123.45, gauge.Value)

		// Ждем асинхронную обработку события фоновым воркером
		select {
		case event := <-observer.events:
			raw, err := json.Marshal(event)
			assert.NoError(t, err)
			assert.Contains(t, string(raw), "192.168.1.50")
			assert.Contains(t, string(raw), "PollCount")
			assert.Contains(t, string(raw), "Alloc")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("expected audit event to be sent")
		}
	})

	t.Run("empty metrics list returns InvalidArgument", func(t *testing.T) {
		_, err := grpcSrv.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty metric id returns InvalidArgument", func(t *testing.T) {
		req := &pb.UpdateMetricsRequest{
			Metrics: []*pb.Metric{
				{Id: "", Type: pb.Metric_COUNTER, Delta: 1},
			},
		}
		_, err := grpcSrv.UpdateMetrics(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}