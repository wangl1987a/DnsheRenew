package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"dnsherene/internal/report"
)

type Webhook struct {
	httpClient *http.Client
}

// Notify 在配置了 webhook 时把结构化结果发送到远端地址。
func (w Webhook) Notify(ctx context.Context, info report.Info) error {
	url := strings.TrimSpace(os.Getenv("DNSHE_NOTIFY_WEBHOOK_URL"))
	token := strings.TrimSpace(os.Getenv("DNSHE_NOTIFY_WEBHOOK_TOKEN"))
	if url == "" {
		return nil
	}

	if info.GeneratedAt.IsZero() {
		info.GeneratedAt = time.Now().UTC()
	}

	raw, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("encode webhook payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build webhook request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	httpClient := w.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
