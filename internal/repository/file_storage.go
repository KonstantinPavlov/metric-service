package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/KonstantinPavlov/metric-service/internal/model"
)

type FileStorage struct {
	mu         sync.RWMutex
	wg         sync.WaitGroup
	cfg        fileStorageCfg
	repository MetricRepository
	log        *zap.Logger
}

type fileStorageCfg struct {
	storeInterval int
	storagePath   string
	restore       bool
}

func NewFileStorage(interval int, path string, restore bool, repository MetricRepository, log *zap.Logger) *FileStorage {
	return &FileStorage{
		cfg: fileStorageCfg{
			storeInterval: interval,
			storagePath:   path,
			restore:       restore,
		},
		repository: repository,
		log:        log,
	}
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (fs *FileStorage) Start(ctx context.Context) error {
	if fs.cfg.restore {
		fs.log.Info("Start restore metrics from storage", zap.String("storage_path", fs.cfg.storagePath))
		fs.Restore(ctx)
		fs.log.Info("End restore metrics from storage", zap.String("storage_path", fs.cfg.storagePath))
	}
	if fs.isSyncSave() {
		return nil
	}
	fs.wg.Add(1)
	go func() {
		defer fs.wg.Done()

		ticker := time.NewTicker(time.Duration(fs.cfg.storeInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fs.log.Info("Start async store metrics...", zap.String("storage_path", fs.cfg.storagePath))
				fs.storeMetricsAsync(ctx)
				fs.log.Info("End async store metrics...", zap.String("storage_path", fs.cfg.storagePath))
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (fs *FileStorage) Stop() {
	fs.wg.Wait()
}

func (fs *FileStorage) Restore(ctx context.Context) {
	data, err := fs.restore()
	if err != nil {
		fs.log.Error("failed to restore data!", zap.Error(err))
	}

	for _, metric := range data {
		fs.log.Info("Restoring metric", zap.String("name", metric.ID), zap.String("metric_type", metric.MType))
		switch metric.MType {
		case model.Counter:
			fs.repository.SaveCounter(ctx, metric.ID, *metric.Delta)
		case model.Gauge:
			fs.repository.SaveGauge(ctx, metric.ID, *metric.Value)
		default:
			fs.log.Warn("Unknown metric type in storage!", zap.String("metric_type", metric.MType))
		}
	}
}

func (fs *FileStorage) storeMetricsAsync(ctx context.Context) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.storeMetrics(ctx)
}

func (fs *FileStorage) storeMetrics(ctx context.Context) {
	data := make([]model.Metrics, 0)
	counters, err := fs.repository.GetNames(ctx, model.Counter)
	if err != nil {
		fs.log.Error("Failed to GetNames for counters!", zap.Error(err))
		return
	}
	for _, counter := range counters {
		metricData, err := fs.repository.GetCounter(ctx, counter)
		if err != nil {
			fs.log.Error("Failed to GetCounter", zap.String("name", counter), zap.Error(err))
			continue
		}
		if metricValue, ok := metricData.Value.(int64); ok {
			metric := model.Metrics{
				ID:    metricData.Name,
				MType: model.Counter,
				Delta: &metricValue,
			}
			data = append(data, metric)
		} else {
			fs.log.Warn("Value type is not int64!", zap.String("name", metricData.Name))
		}
	}
	gauges, err := fs.repository.GetNames(ctx, model.Gauge)
	if err != nil {
		fs.log.Error("Failed to GetNames for gauges!", zap.Error(err))
		return
	}
	for _, gauge := range gauges {
		metricData, err := fs.repository.GetGauge(ctx, gauge)
		if err != nil {
			fs.log.Error("Failed to GetGauge", zap.String("name", gauge), zap.Error(err))
			continue
		}
		if metricValue, ok := metricData.Value.(float64); ok {
			metric := model.Metrics{
				ID:    metricData.Name,
				MType: model.Gauge,
				Value: &metricValue,
			}
			data = append(data, metric)
		} else {
			fs.log.Warn("Value type is not float64!", zap.String("name", metricData.Name))
		}
	}

	err = fs.store(data)
	if err != nil {
		fs.log.Error("Failed to store metrics", zap.Error(err))
	} else {
		fs.log.Info("Stored metrics", zap.String("storage_path", fs.cfg.storagePath), zap.Int("count", len(data)))
	}
}

func (fs *FileStorage) store(data any) error {
	file, err := os.Create(fs.cfg.storagePath)
	if err != nil {
		return fmt.Errorf("Failed to create/open file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("Failed to encode to json: %w", err)
	}
	return nil
}

func (fs *FileStorage) restore() ([]model.Metrics, error) {
	var data []model.Metrics
	file, err := os.Open(fs.cfg.storagePath)
	if err != nil {
		return data, fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("Error decode data: %w", err)
	}
	return data, nil
}

func (fs *FileStorage) GetNames(ctx context.Context, metricType string) ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.repository.GetNames(ctx, metricType)
}

func (fs *FileStorage) GetCounter(ctx context.Context, name string) (*MetricData, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.repository.GetCounter(ctx, name)
}

func (fs *FileStorage) GetGauge(ctx context.Context, name string) (*MetricData, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.repository.GetGauge(ctx, name)
}

func (fs *FileStorage) SaveCounter(ctx context.Context, name string, value int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.isSyncSave() {
		res := fs.repository.SaveCounter(ctx, name, value)
		fs.storeMetrics(ctx)
		return res
	}
	return fs.repository.SaveCounter(ctx, name, value)
}

func (fs *FileStorage) SaveGauge(ctx context.Context, name string, value float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.isSyncSave() {
		res := fs.repository.SaveGauge(ctx, name, value)
		fs.storeMetrics(ctx)
		return res
	}
	return fs.repository.SaveGauge(ctx, name, value)
}

func (fs *FileStorage) SaveMetrics(
	ctx context.Context,
	counters []MetricData,
	gauges []MetricData,
) error {
	for _, counter := range counters {
		if metricValue, ok := counter.Value.(int64); ok {
			fs.SaveCounter(ctx, counter.Name, metricValue)
		} else {
			return fmt.Errorf("Value is not a int64!")
		}
	}
	for _, counter := range gauges {
		if metricValue, ok := counter.Value.(float64); ok {
			fs.SaveGauge(ctx, counter.Name, metricValue)
		} else {
			return fmt.Errorf("Value is not a float64!")
		}
	}
	return nil
}

func (fs *FileStorage) isSyncSave() bool {
	return fs.cfg.storeInterval <= 0
}
