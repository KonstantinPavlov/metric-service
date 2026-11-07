package middleware

import (
	"context"
	"net"
	"testing"

	pb "github.com/KonstantinPavlov/metric-service/internal/proto"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type stubMetricsServer struct {
	pb.UnimplementedMetricsServer
}

func (s *stubMetricsServer) UpdateMetrics(_ context.Context, _ *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	return &pb.UpdateMetricsResponse{}, nil
}

func TestGRPCSubnetInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		trustedCIDR  string
		clientIP     string
		sendMetadata bool
		wantCode     codes.Code
	}{
		{
			name:         "no subnet restriction allows any request",
			trustedCIDR:  "",
			sendMetadata: false,
			wantCode:     codes.OK,
		},
		{
			name:         "ip inside trusted subnet is allowed",
			trustedCIDR:  "192.168.1.0/24",
			clientIP:     "192.168.1.42",
			sendMetadata: true,
			wantCode:     codes.OK,
		},
		{
			name:         "ip outside trusted subnet is denied",
			trustedCIDR:  "192.168.1.0/24",
			clientIP:     "10.0.0.5",
			sendMetadata: true,
			wantCode:     codes.PermissionDenied,
		},
		{
			name:         "missing metadata is denied when subnet is configured",
			trustedCIDR:  "192.168.1.0/24",
			sendMetadata: false,
			wantCode:     codes.PermissionDenied,
		},
		{
			name:         "invalid ip string is denied",
			trustedCIDR:  "192.168.1.0/24",
			clientIP:     "not-an-ip",
			sendMetadata: true,
			wantCode:     codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor, err := GRPCSubnetInterceptor(tt.trustedCIDR)
			assert.NoError(t, err)

			// Создаем виртуальное сетевое соединение в памяти (bufconn)
			lis := bufconn.Listen(1024 * 1024)
			srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
			pb.RegisterMetricsServer(srv, &stubMetricsServer{})

			go func() {
				_ = srv.Serve(lis)
			}()
			defer srv.Stop()

			// Настраиваем клиент на работу через буфер в памяти
			conn, err := grpc.NewClient(
				"passthrough:///bufnet",
				grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return lis.DialContext(ctx)
				}),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			assert.NoError(t, err)
			defer conn.Close()

			client := pb.NewMetricsClient(conn)

			ctx := context.Background()
			if tt.sendMetadata {
				md := metadata.Pairs("x-real-ip", tt.clientIP)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}

			_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{})
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
