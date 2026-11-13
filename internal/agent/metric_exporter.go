package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/KonstantinPavlov/metric-service/internal/config"
	"github.com/KonstantinPavlov/metric-service/internal/crypto"
	"github.com/KonstantinPavlov/metric-service/internal/model"
	pb "github.com/KonstantinPavlov/metric-service/internal/proto"
	"github.com/KonstantinPavlov/metric-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type MetricsExporter struct {
	serverURL string
	// grpc
	grpcAddr   string
	grpcConn   *grpc.ClientConn
	grpcClient pb.MetricsClient
	//
	provider        service.MetricsProvider
	client          http.Client
	log             *zap.Logger
	cryptoKeyString string
	cryptoKey       []byte
	wg              sync.WaitGroup
	rateLimit       int
	jobsChannel     chan []model.Metrics
	// for jobsChannel close
	closeOnce sync.Once
}

func NewMetricsExporter(serverURL string, grpcAddr string, provider service.MetricsProvider, client http.Client, log *zap.Logger, cryptoKey string, publicCryptoKeyPath string, rateLimit int) (*MetricsExporter, error) {
	if publicCryptoKeyPath != "" {
		encryptTransport, err := crypto.NewAsymmetricEncryptTransport(publicCryptoKeyPath, client.Transport)
		if err != nil {
			log.Error("Failed to initialize asymmetric crypto transport", zap.Error(err))
			return nil, err
		}
		client.Transport = encryptTransport
		log.Info("Asymmetric encryption layer added successfully", zap.String("path", publicCryptoKeyPath))

	}

	var grpcConn *grpc.ClientConn
	var grpcClient pb.MetricsClient
	if grpcAddr != "" {
		var err error
		grpcConn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpcClient = pb.NewMetricsClient(grpcConn)
		log.Info("gRPC client initialized", zap.String("addr", grpcAddr))
	}

	return &MetricsExporter{
		serverURL:       serverURL,
		grpcAddr:        grpcAddr,
		grpcConn:        grpcConn,
		grpcClient:      grpcClient,
		provider:        provider,
		client:          client,
		log:             log,
		cryptoKeyString: cryptoKey,
		cryptoKey:       make([]byte, 0),
		rateLimit:       rateLimit,
		jobsChannel:     make(chan []model.Metrics, rateLimit*2),
	}, nil
}

func (me *MetricsExporter) Start(ctx context.Context, interval time.Duration) error {
	if me.cryptoKeyString != "" {
		keyBytes, err := crypto.DecodeKeyString(me.cryptoKeyString)
		if err != nil {
			return err
		}
		me.log.Info("Success load key!")
		me.cryptoKey = keyBytes
	}

	// запускаем воркеров
	for i := 0; i < me.rateLimit; i++ {
		me.wg.Add(1)
		go me.worker(ctx)
	}

	me.wg.Add(1)
	go func() {
		defer me.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				me.log.Info("Start exporting metrics...")
				me.Export(ctx)
				me.log.Info("End exporting metrics...")
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (me *MetricsExporter) Stop(ctx context.Context) error {
	me.log.Info("Forcing final metrics export (flush) before exit...")
	// final export
	me.Export(ctx)
	// closing jobs channel here now
	me.closeOnce.Do(func() {
		close(me.jobsChannel)
	})

	c := make(chan struct{})
	go func() {
		me.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		if me.grpcConn != nil {
			_ = me.grpcConn.Close()
		}
		me.log.Info("All background workers stopped cleanly.")
		return nil

	case <-ctx.Done():
		if me.grpcConn != nil {
			_ = me.grpcConn.Close()
		}
		return ctx.Err()
	}
}

func (me *MetricsExporter) Export(ctx context.Context) {
	requests := make([]model.Metrics, 0)
	for key, value := range me.provider.GetCounters(ctx) {
		requests = append(requests, model.Metrics{
			ID:    key,
			MType: model.Counter,
			Delta: &value,
		})
	}
	for key, value := range me.provider.GetGauges(ctx) {
		requests = append(requests, model.Metrics{
			ID:    key,
			MType: model.Gauge,
			Value: &value,
		})
	}
	if len(requests) == 0 {
		me.log.Warn("Metrics are empty!")
		return
	}

	select {
	case me.jobsChannel <- requests:
	case <-ctx.Done():
		me.log.Warn("Skipped sending metrics to channel due to context expiration")
		return
	default:
		me.log.Warn("Jobs channel is no longer accepting metrics")
	}
}

func (me *MetricsExporter) worker(ctx context.Context) {
	defer me.wg.Done()

	for {
		select {
		case metrics, ok := <-me.jobsChannel:
			if !ok {
				return
			}
			if me.grpcClient != nil {
				me.postMetricsGRPC(ctx, metrics)
			} else {
				me.postMetrics(ctx, metrics)
			}

		case <-ctx.Done():
			// gracefull shutdown
			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			for metrics := range me.jobsChannel {
				if err := flushCtx.Err(); err != nil {
					me.log.Error("Failing to flush remaining metrics: shutdown timeout reached", zap.Error(err))
					cancel()
					return
				}
				if me.grpcClient != nil {
					me.postMetricsGRPC(flushCtx, metrics)
				} else {
					me.postMetrics(flushCtx, metrics)
				}
			}
			cancel()
			return
		}
	}
}

func (me *MetricsExporter) postMetrics(ctx context.Context, requests []model.Metrics) {
	jsonBytes, err := json.Marshal(requests)
	if err != nil {
		me.log.Error("Failed to marshall request!", zap.Error(err))
		return
	}
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(jsonBytes); err != nil {
		me.log.Error("Failed to compress data!", zap.Error(err))
		return
	}
	if err := gzWriter.Close(); err != nil {
		me.log.Error("Failed to close compress data!", zap.Error(err))
		return
	}

	me.log.Debug("Compressed data", zap.Int("before", len(jsonBytes)), zap.Int("after", compressed.Len()))
	compressedBytes := compressed.Bytes()
	maxAttemps := 4
	retryDelay := 1 * time.Second

	var resp *http.Response

	var sha256HeaderValue = ""
	if len(me.cryptoKey) != 0 {
		sha256HeaderValue = crypto.CalculateSha256(me.cryptoKey, compressedBytes)
	}

	for attempt := 1; attempt <= maxAttemps; attempt++ {
		if err := ctx.Err(); err != nil {
			me.log.Error("Context cancelled during HTTP retries", zap.Error(err))
			return
		}
		bodyReader := bytes.NewReader(compressedBytes)
		request, err := http.NewRequestWithContext(ctx, "POST", "http://"+me.serverURL+"/updates/", bodyReader)
		if err != nil {
			me.log.Error("Failed to create request!", zap.Error(err))
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		if sha256HeaderValue != "" {
			request.Header.Set("HashSHA256", sha256HeaderValue)
		}
		request.Header.Set("X-Real-IP", config.GetMachineIP())
		resp, err = me.client.Do(request)
		if err == nil {
			// all fine!
			break
		}
		if attempt == maxAttemps {
			me.log.Error("Error publishing metrics after all retries", zap.Int("size", len(requests)), zap.Error(err))
			return
		}

		me.log.Warn("Network error occurred, retrying...", zap.Int("attempt", attempt), zap.Error(err))

		select {
		case <-time.After(retryDelay):
			retryDelay += 2 * time.Second
		case <-ctx.Done():
			me.log.Error("Context cancelled while waiting for next retry", zap.Error(ctx.Err()))
			return
		}
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp != nil && resp.Header.Get("Content-Type") != "application/json" {
		me.log.Error("Server response content-type is not valid!")
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			me.log.Error("Error reading response body", zap.Error(err))
			return
		}
		me.log.Info("Server response body", zap.String("body", string(bodyBytes)))
	}
}


func (me *MetricsExporter) postMetricsGRPC(ctx context.Context, requests []model.Metrics) {
	pbMetrics := make([]*pb.Metric, 0, len(requests))
	for _, r := range requests {
		m := &pb.Metric{Id: r.ID}
		switch r.MType {
		case model.Counter:
			m.Type = pb.Metric_COUNTER
			if r.Delta != nil {
				m.Delta = *r.Delta
			}
		case model.Gauge:
			m.Type = pb.Metric_GAUGE
			if r.Value != nil {
				m.Value = *r.Value
			}
		}
		pbMetrics = append(pbMetrics, m)
	}

	req := &pb.UpdateMetricsRequest{Metrics: pbMetrics}
	md := metadata.Pairs("x-real-ip", config.GetMachineIP())

	maxAttempts := 4
	retryDelay := 1 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			me.log.Error("Context cancelled during gRPC retries", zap.Error(err))
			return
		}

		reqCtx := metadata.NewOutgoingContext(ctx, md)
		_, err := me.grpcClient.UpdateMetrics(reqCtx, req)
		if err == nil {
			return
		}

		if attempt == maxAttempts {
			me.log.Error("Error publishing metrics via gRPC after all retries", zap.Int("size", len(requests)), zap.Error(err))
			return
		}

		me.log.Warn("gRPC error occurred, retrying...", zap.Int("attempt", attempt), zap.Error(err))

		select {
		case <-time.After(retryDelay):
			retryDelay += 2 * time.Second
		case <-ctx.Done():
			me.log.Error("Context cancelled while waiting for next gRPC retry", zap.Error(ctx.Err()))
			return
		}
	}
}