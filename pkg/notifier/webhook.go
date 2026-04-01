package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Webhook struct {
	url        string
	token      string
	httpClient *http.Client
}

func NewWebhook(url string, token string, httpClient *http.Client) (*Webhook, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Webhook{
		url:        url,
		token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}, nil
}

func (w *Webhook) Notify(ctx context.Context, event Event) error {
	// 使用统一事件结构序列化，保证不同通知通道字段语义一致。
	payload := map[string]any{
		"level":   event.Level,
		"title":   event.Title,
		"message": event.Message,
		"fields":  event.Fields,
		"time":    event.Time.UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode webhook payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build webhook request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}

	resp, err := w.httpClient.Do(req)
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
