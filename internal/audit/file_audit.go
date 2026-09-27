package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileAuditObserver struct {
	mu       sync.Mutex
	log      *zap.Logger
	filePath string
}

func NewAuditFileObserver(log *zap.Logger, filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		log:      log,
		filePath: filePath,
	}
}

func (a *FileAuditObserver) Notify(
	ctx context.Context,
	event AuditEvent,
) {
	if err := ctx.Err(); err != nil {
		a.log.Warn("Skipping audit log: context already done", zap.Error(err))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		a.log.Error("Failed to open file", zap.Error(err))
		return
	}
	defer file.Close()
	a.log.Info("Saving audit event to file ", zap.Any("event", event))
	if err := json.NewEncoder(file).Encode(event); err != nil {
		a.log.Error("Failed to encode to json", zap.Error(err))
		return
	}
}
