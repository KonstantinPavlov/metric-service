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

func (fs *FileStorage) Start(ctx context.Context) {
	if fs.cfg.restore {
		fs.log.Info("Start restore metrics from storage", zap.String("storage_path", fs.cfg.storagePath))
		fs.Restore()
		fs.log.Info("End restore metrics from storage", zap.String("storage_path", fs.cfg.storagePath))
	}
	if fs.isSyncSave() {
		return
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
				fs.StoreMetrics()
				fs.log.Info("End async store metrics...", zap.String("storage_path", fs.cfg.storagePath))
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (fs *FileStorage) Stop() {
	fs.wg.Wait()
}

func (fs *FileStorage) Restore() {
	data, err := fs.restore()
	if err != nil {
		fs.log.Error("failed to restore data!", zap.Error(err))
	}

	for _, metric := range data {
		fs.log.Info("Restoring metric", zap.String("name", metric.ID), zap.String("metric_type", metric.MType))
		switch metric.MType {
		case model.Counter:
			fs.repository.SaveCounter(metric.ID, *metric.Delta)
		case model.Gauge:
			fs.repository.SaveGauge(metric.ID, *metric.Value)
		default:
			fs.log.Warn("Unknown metric type in storage!", zap.String("metric_type", metric.MType))
		}
	}
}

func (fs *FileStorage) StoreMetrics() {
	data := make([]model.Metrics, 0)
	for _, counter := range fs.repository.GetNames(model.Counter) {
		metricData := fs.repository.GetCounter(counter)
		metricValue := metricData.Value.(int64)
		metric := model.Metrics{
			ID:    metricData.Name,
			MType: model.Counter,
			Delta: &metricValue,
		}
		data = append(data, metric)
	}

	for _, gauge := range fs.repository.GetNames(model.Gauge) {
		metricData := fs.repository.GetGauge(gauge)
		metricValue := metricData.Value.(float64)
		metric := model.Metrics{
			ID:    metricData.Name,
			MType: model.Gauge,
			Value: &metricValue,
		}
		data = append(data, metric)
	}

	err := fs.store(data)
	if err != nil {
		fs.log.Error("Failed to store metrics", zap.Error(err))
	}
	fs.log.Info("Stored metrics", zap.String("storage_path", fs.cfg.storagePath), zap.Int("count", len(data)))
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

func (fs *FileStorage) GetNames(metricType string) []string {
	return fs.repository.GetNames(metricType)
}

func (fs *FileStorage) GetCounter(name string) *MetricData {
	return fs.repository.GetCounter(name)
}

func (fs *FileStorage) GetGauge(name string) *MetricData {
	return fs.repository.GetGauge(name)
}

func (fs *FileStorage) SaveCounter(name string, value int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.isSyncSave() {
		res := fs.repository.SaveCounter(name, value)
		fs.StoreMetrics()
		return res
	}
	return fs.repository.SaveCounter(name, value)
}

func (fs *FileStorage) SaveGauge(name string, value float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.isSyncSave() {
		res := fs.repository.SaveGauge(name, value)
		fs.StoreMetrics()
		return res
	}
	return fs.repository.SaveGauge(name, value)
}

func (fs *FileStorage) isSyncSave() bool {
	return fs.cfg.storeInterval <= 0
}
