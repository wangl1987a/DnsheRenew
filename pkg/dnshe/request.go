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
	Error     string `json:"error"`
	Message   string `json:"message"`
	Limit     *int   `json:"limit"`
	Remaining *int   `json:"remaining"`
	ResetAt   string `json:"reset_at"`
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
		return decodeAPIError(resp.StatusCode, respBody)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response failed: %w; body=%s", err, truncate(respBody, 512))
	}
	return nil
}

// ensureSuccess 校验业务响应中的 success 字段并转换为结构化错误。
func ensureSuccess(operation string, resp baseResponse) error {
	if resp.Success {
		return nil
	}
	return &APIError{
		Operation: operation,
		ErrorText: strings.TrimSpace(resp.Error),
		Message:   strings.TrimSpace(resp.Message),
	}
}

// decodeAPIError 将 HTTP 错误响应体解析为 APIError。
func decodeAPIError(statusCode int, body []byte) error {
	errResp := &APIError{
		StatusCode: statusCode,
		RawBody:    truncate(body, 256),
	}

	var payload apiError
	if err := json.Unmarshal(body, &payload); err == nil {
		apiErr := &APIError{
			StatusCode: statusCode,
			ErrorText:  strings.TrimSpace(payload.Error),
			Message:    strings.TrimSpace(payload.Message),
			Limit:      payload.Limit,
			Remaining:  payload.Remaining,
			ResetAt:    strings.TrimSpace(payload.ResetAt),
			RawBody:    truncate(body, 256),
		}
		return apiErr
	}

	if len(body) == 0 {
		errResp.RawBody = "empty response"
	}
	return errResp
}

// truncate 截断过长响应体，避免错误信息无限增长。
func truncate(in []byte, max int) string {
	s := strings.TrimSpace(string(in))
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// firstNonEmpty 返回第一个去空格后非空的字符串。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
