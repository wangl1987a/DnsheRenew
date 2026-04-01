package dnshe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type baseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// requestJSON 是统一 HTTP 调用入口，负责公共 query/header 与通用错误处理。
func (c *Client) requestJSON(
	ctx context.Context,
	method string,
	endpoint string,
	action string,
	query url.Values,
	body any,
	out any,
) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("m", "domain_hub")
	q.Set("endpoint", endpoint)
	if action != "" {
		q.Set("action", action)
	}
	for key, values := range query {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	u.RawQuery = q.Encode()

	var bodyReader io.Reader
	if body != nil {
		raw, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return fmt.Errorf("marshal request body failed: %w", marshalErr)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-API-Secret", c.apiSecret)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, decodeAPIError(respBody))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response failed: %w; body=%s", err, truncate(respBody, 512))
	}
	return nil
}

func ensureSuccess(operation string, resp baseResponse) error {
	if resp.Success {
		return nil
	}
	return fmt.Errorf("%s failed: %s", operation, firstNonEmpty(resp.Error, resp.Message, "unknown error"))
}

func decodeAPIError(body []byte) string {
	var out apiError
	if err := json.Unmarshal(body, &out); err == nil {
		if msg := firstNonEmpty(out.Error, out.Message); msg != "" {
			return msg
		}
	}

	if len(body) == 0 {
		return "empty response"
	}
	return truncate(body, 256)
}

func truncate(in []byte, max int) string {
	s := strings.TrimSpace(string(in))
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
