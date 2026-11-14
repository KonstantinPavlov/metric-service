package handler

import (
	"context"

	"github.com/KonstantinPavlov/metric-service/internal/audit"
	pb "github.com/KonstantinPavlov/metric-service/internal/proto"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCMetricsServer struct {
	pb.UnimplementedMetricsServer
	handler *MetricHandler
}

func NewGRPCMetricsServer(handler *MetricHandler) *GRPCMetricsServer {
	return &GRPCMetricsServer{handler: handler}
}

func (s *GRPCMetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if len(req.Metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "received empty metrics list")
	}

	var countersData []repository.MetricData
	var gaugesData []repository.MetricData
	var metricNames []string

	for _, m := range req.Metrics {
		if m.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "metric id must be set")
		}

		switch m.Type {
		case pb.Metric_COUNTER:
			metricNames = append(metricNames, m.Id)
			countersData = append(countersData, repository.MetricData{
				Name:  m.Id,
				Value: m.Delta,
			})
		case pb.Metric_GAUGE:
			metricNames = append(metricNames, m.Id)
			gaugesData = append(gaugesData, repository.MetricData{
				Name:  m.Id,
				Value: m.Value,
			})
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", m.Type)
		}
	}

	if err := s.handler.Repository.SaveMetrics(ctx, countersData, gaugesData); err != nil {
		s.handler.log.Error("Failed to save metrics via gRPC", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to save metrics")
	}

	clientIP := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-real-ip"); len(ips) > 0 {
			clientIP = ips[0]
		}
	}

	s.handler.notify(ctx, audit.NewServerEvent(metricNames, clientIP))

	return &pb.UpdateMetricsResponse{}, nil
}
