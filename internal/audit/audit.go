package audit

import (
	"context"
	"time"
)

type AuditEvent struct {
	Time    int64    `json:"ts"`
	Metrics []string `json:"metrics"`
	Address string   `json:"ip_address"`
}

func NewServerEvent(metrics []string, address string) AuditEvent {
	return AuditEvent{
		Time:    time.Now().Unix(),
		Metrics: metrics,
		Address: address,
	}
}

type AuditObserver interface {
	Notify(
		ctx context.Context,
		event AuditEvent,
	)
	Stop()
}
