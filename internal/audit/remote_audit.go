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

func (r *RemoteAuditObserver) Notify(
	ctx context.Context,
	event AuditEvent,
) {

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(event); err != nil {
		r.log.Error("Failed to encode audit event to JSON", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.serverURL, &buf)
	if err != nil {
		r.log.Error("Failed to create HTTP request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		r.log.Error("Failed to send audit event via HTTP", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.log.Error("Remote audit server returned non-2xx status",
			zap.Int("status_code", resp.StatusCode),
		)
		return
	}
	r.log.Info("Audit data successfully sent to remote server")
}

func (r *RemoteAuditObserver) Stop() {
	//do nothing
}
