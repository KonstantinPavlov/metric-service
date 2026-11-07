package agent

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	pb "github.com/KonstantinPavlov/metric-service/internal/proto"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"github.com/KonstantinPavlov/metric-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestMetricsExporter_Start(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	calledChan := make(chan struct{})
	mock := &testutils.MockProvider{CalledChan: calledChan}

	mc, err := NewMetricsExporter(
		"localhost:8080",
		"", // grpcAddr
		mock,
		http.Client{},
		zapLogger,
		"",
		"",
		1,
	)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = mc.Start(ctx, 10*time.Millisecond)
	assert.NoError(t, err)

	for i := 1; i <= 2; i++ {
		select {
		case <-calledChan:
			// do nothing - was called
		case <-time.After(50 * time.Millisecond):
			assert.Failf(t, "Timeout", "Metric collection not executed at tick №%d", i)
			return
		}
	}
}

func TestMetricsExporter_Stop(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	calledChan := make(chan struct{}, 10)
	mock := &testutils.MockProvider{CalledChan: calledChan}

	mc, err := NewMetricsExporter(
		"localhost:8080",
		"", // grpcAddr
		mock,
		http.Client{},
		zapLogger,
		"",
		"",
		1,
	)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	err = mc.Start(ctx, 10*time.Millisecond)
	assert.NoError(t, err)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer shutdownCancel()
	err = mc.Stop(shutdownCtx)
	assert.NoError(t, err, "Exporter must stop cleanly within shutdown context timeout")
}

type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMetricsExporter_Export(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	callCounter := 0
	storage := repository.NewMemStorage()
	provider := service.NewDefaultProvider(storage, zapLogger)
	provider.SaveCounter(t.Context(), "some-counter", 1)
	provider.SaveGauge(t.Context(), "some-gauge", 1)
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	exporter, err := NewMetricsExporter(
		"",
		"", // grpcAddr
		provider,
		http.Client{
			Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				callCounter++
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			}),
		},
		zapLogger,
		"",
		"",
		1,
	)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = exporter.Start(ctx, 10*time.Hour)
	assert.NoError(t, err)
	exporter.Export(context.Background())

	assert.Eventually(t, func() bool {
		return callCounter == 1
	}, 500*time.Millisecond, 10*time.Millisecond, "Export Method calling http.Client.Post 1 times")
}

// mockMetricsServer реализует pb.MetricsServer для тестирования gRPC-клиента агента
type mockMetricsServer struct {
	pb.UnimplementedMetricsServer
	mu          sync.Mutex
	receivedReq *pb.UpdateMetricsRequest
	receivedIP  string
}

func (m *mockMetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivedReq = req
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-real-ip"); len(ips) > 0 {
			m.receivedIP = ips[0]
		}
	}
	return &pb.UpdateMetricsResponse{}, nil
}

func TestMetricsExporter_ExportGRPC(t *testing.T) {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	// Поднимаем in-memory gRPC сервер через bufconn
	lis := bufconn.Listen(1024 * 1024)
	grpcSrv := grpc.NewServer()
	mockSrv := &mockMetricsServer{}
	pb.RegisterMetricsServer(grpcSrv, mockSrv)

	go func() {
		_ = grpcSrv.Serve(lis)
	}()
	defer grpcSrv.Stop()

	// Подключаем клиент к bufconn
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	storage := repository.NewMemStorage()
	provider := service.NewDefaultProvider(storage, zapLogger)
	provider.SaveCounter(t.Context(), "grpc-counter", 42)
	provider.SaveGauge(t.Context(), "grpc-gauge", 3.14)

	exporter, err := NewMetricsExporter(
		"",
		"", // Оставляем пустым, чтобы не создавать реальный TCP-сокет, и подменяем клиент ниже
		provider,
		http.Client{},
		zapLogger,
		"",
		"",
		1,
	)
	assert.NoError(t, err)
	exporter.grpcConn = conn
	exporter.grpcClient = pb.NewMetricsClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = exporter.Start(ctx, 10*time.Hour)
	assert.NoError(t, err)

	exporter.Export(context.Background())

	assert.Eventually(t, func() bool {
		mockSrv.mu.Lock()
		defer mockSrv.mu.Unlock()
		return mockSrv.receivedReq != nil
	}, 500*time.Millisecond, 10*time.Millisecond, "gRPC server must receive UpdateMetricsRequest")

	mockSrv.mu.Lock()
	defer mockSrv.mu.Unlock()
	assert.Len(t, mockSrv.receivedReq.Metrics, 2)
	assert.NotEmpty(t, mockSrv.receivedIP, "x-real-ip metadata must be passed")
}
