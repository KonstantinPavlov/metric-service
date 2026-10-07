package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileAuditObserver struct {
	mu   sync.Mutex
	log  *zap.Logger
	file *os.File
}

func NewAuditFileObserver(log *zap.Logger, filePath string) (*FileAuditObserver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Failed to open file", zap.Error(err))
		return nil, err
	}
	return &FileAuditObserver{
		log:  log,
		file: file,
	}, nil
}

func (f *FileAuditObserver) Notify(
	ctx context.Context,
	event AuditEvent,
) {
	if err := ctx.Err(); err != nil {
		f.log.Warn("Skipping audit log: context already done", zap.Error(err))
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.log.Info("Saving audit event to file ", zap.Any("event", event))
	if err := json.NewEncoder(f.file).Encode(event); err != nil {
		f.log.Error("Failed to encode to json", zap.Error(err))
		return
	}
}

func (f *FileAuditObserver) Stop() {
	f.file.Close()
}
