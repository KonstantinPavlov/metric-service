package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAuditFileObserver_Notify_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.log")
	observer, err := NewAuditFileObserver(logger, filePath)
	if err != nil {
		t.Error(err)
	}
	event1 := NewServerEvent([]string{"cpu_usage", "ram_usage"}, "192.168.1.1")
	time.Sleep(10 * time.Millisecond)
	event2 := NewServerEvent([]string{"disk_write"}, "10.0.0.5")

	observer.Notify(t.Context(), event1)
	observer.Notify(t.Context(), event2)

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open audit file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var result1 AuditEvent
	if err := decoder.Decode(&result1); err != nil {
		t.Fatalf("failed to decode first event: %v", err)
	}
	if result1.Time != event1.Time {
		t.Errorf("expected time %v, got %v", event1.Time, result1.Time)
	}
	if result1.Address != event1.Address {
		t.Errorf("expected address %s, got %s", event1.Address, result1.Address)
	}
	if len(result1.Metrics) != len(event1.Metrics) || result1.Metrics[0] != event1.Metrics[0] {
		t.Errorf("expected metrics %v, got %v", event1.Metrics, result1.Metrics)
	}

	var result2 AuditEvent
	if err := decoder.Decode(&result2); err != nil {
		t.Fatalf("failed to decode second event: %v", err)
	}
	if result2.Time != event2.Time {
		t.Errorf("expected time %v, got %v", event2.Time, result2.Time)
	}
	if result2.Address != event2.Address {
		t.Errorf("expected address %s, got %s", event2.Address, result2.Address)
	}

	var result3 AuditEvent
	if err := decoder.Decode(&result3); err != io.EOF {
		t.Errorf("expected EOF, got error: %v", err)
	}
}

func TestAuditFileObserver_Notify_ContextCancelled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit.log")

	observer, err := NewAuditFileObserver(logger, filePath)
	if err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := NewServerEvent([]string{"network_in"}, "127.0.0.1")
	observer.Notify(ctx, event)

	info, err := os.Stat(filePath)
	if err == nil && info.Size() > 0 {
		t.Error("file should be empty or not exist when context is cancelled")
	}
}

func TestAuditFileObserver_Notify_Concurrency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "audit_concurrent.log")

	observer, err := NewAuditFileObserver(logger, filePath)
	if err != nil {
		t.Error(err)
	}
	const goroutinesCount = 50
	var wg sync.WaitGroup
	wg.Add(goroutinesCount)

	for i := 0; i < goroutinesCount; i++ {
		go func(idx int) {
			defer wg.Done()
			event := NewServerEvent([]string{"metric_load"}, "192.168.1."+fmt.Sprint(idx))
			observer.Notify(context.Background(), event)
		}(i)
	}
	wg.Wait()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open audit file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	linesCount := 0

	for {
		var e AuditEvent
		if err := decoder.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("failed to decode JSON line: %v. JSON might be corrupted!", err)
		}
		linesCount++
	}

	if linesCount != goroutinesCount {
		t.Errorf("expected %d records, but got %d", goroutinesCount, linesCount)
	}
}
