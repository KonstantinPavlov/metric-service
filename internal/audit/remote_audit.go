package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type RemoteAuditObserver struct {
	log       *zap.Logger
	serverURL string
	client    http.Client
}

func NewRemoteAuditObserver(log *zap.Logger, url string) *RemoteAuditObserver {
	return &RemoteAuditObserver{
		log:       log,
		serverURL: url,
		client:    http.Client{},
	}
}

func (a *RemoteAuditObserver) Notify(
	ctx context.Context,
	event AuditEvent,
) {

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(event); err != nil {
		a.log.Error("Failed to encode audit event to JSON", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.serverURL, &buf)
	if err != nil {
		a.log.Error("Failed to create HTTP request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		a.log.Error("Failed to send audit event via HTTP", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.log.Error("Remote audit server returned non-2xx status",
			zap.Int("status_code", resp.StatusCode),
		)
		return
	}
	a.log.Info("Audit data successfully sent to remote server")
}
