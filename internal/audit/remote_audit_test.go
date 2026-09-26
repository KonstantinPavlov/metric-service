package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestRemoteAuditObserver_Notify_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	event := NewServerEvent([]string{"cpu_usage", "ram_usage"}, "192.168.1.1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var received AuditEvent
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if !received.Time.Equal(event.Time) {
			t.Errorf("expected time %v, got %v", event.Time, received.Time)
		}
		if received.Address != event.Address {
			t.Errorf("expected address %s, got %s", event.Address, received.Address)
		}
		if len(received.Metrics) != len(event.Metrics) || received.Metrics[0] != event.Metrics[0] {
			t.Errorf("expected metrics %v, got %v", event.Metrics, received.Metrics)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	observer := NewRemoteAuditObserver(logger, server.URL)
	observer.Notify(t.Context(), event)
}

func TestRemoteAuditObserver_Notify_ContextCancelled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	event := NewServerEvent([]string{"network_in"}, "127.0.0.1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not receive any requests because context is already cancelled")
	}))
	defer server.Close()
	observer := NewRemoteAuditObserver(logger, server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	observer.Notify(ctx, event)
}

func TestRemoteAuditObserver_Notify_Concurrency(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	const goroutinesCount = 50
	var mu sync.Mutex
	receivedCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received AuditEvent
		_ = json.NewDecoder(r.Body).Decode(&received)

		mu.Lock()
		receivedCount++
		mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	observer := NewRemoteAuditObserver(logger, server.URL)

	var wg sync.WaitGroup
	wg.Add(goroutinesCount)

	for i := 0; i < goroutinesCount; i++ {
		go func(idx int) {
			defer wg.Done()
			event := NewServerEvent([]string{"metric_load"}, "10.0.0.1")
			observer.Notify(context.Background(), event)
		}(i)
	}
	wg.Wait()

	if receivedCount != goroutinesCount {
		t.Errorf("expected server to receive %d requests, but got %d", goroutinesCount, receivedCount)
	}
}
