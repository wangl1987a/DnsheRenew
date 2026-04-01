package dnshe

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type listAPIKeysResponse struct {
	baseResponse
	Count int      `json:"count"`
	Keys  []APIKey `json:"keys"`
}

type createAPIKeyResponse struct {
	baseResponse
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Warning   string `json:"warning"`
}

type regenerateAPIKeyResponse struct {
	baseResponse
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Warning   string `json:"warning"`
}

// ListAPIKeys 获取当前账号全部 API Key 列表。
func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	var out listAPIKeysResponse
	if err := c.requestJSON(ctx, http.MethodGet, "keys", "list", nil, nil, &out); err != nil {
		return nil, err
	}
	if err := ensureSuccess("list api keys", out.baseResponse); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

// CreateAPIKey 创建新的 API Key。
func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (CreateAPIKeyResult, error) {
	keyName := strings.TrimSpace(req.KeyName)
	if keyName == "" {
		return CreateAPIKeyResult{}, fmt.Errorf("keyName is required")
	}

	payload := map[string]any{
		"key_name": keyName,
	}
	if whitelist := strings.TrimSpace(req.IPWhitelist); whitelist != "" {
		payload["ip_whitelist"] = whitelist
	}

	var out createAPIKeyResponse
	if err := c.requestJSON(ctx, http.MethodPost, "keys", "create", nil, payload, &out); err != nil {
		return CreateAPIKeyResult{}, err
	}
	if err := ensureSuccess("create api key", out.baseResponse); err != nil {
		return CreateAPIKeyResult{}, err
	}

	return CreateAPIKeyResult{
		APIKey:    out.APIKey,
		APISecret: out.APISecret,
		Warning:   strings.TrimSpace(out.Warning),
		Message:   strings.TrimSpace(out.Message),
	}, nil
}

// DeleteAPIKey 删除指定 API Key。
func (c *Client) DeleteAPIKey(ctx context.Context, keyID int) error {
	if keyID <= 0 {
		return fmt.Errorf("keyID must be positive")
	}

	payload := map[string]any{
		"key_id": keyID,
	}

	var out baseResponse
	if err := c.requestJSON(ctx, http.MethodPost, "keys", "delete", nil, payload, &out); err != nil {
		return err
	}
	return ensureSuccess("delete api key", out)
}

// RegenerateAPIKey 重置指定 API Key 的 secret。
func (c *Client) RegenerateAPIKey(ctx context.Context, keyID int) (RegenerateAPIKeyResult, error) {
	if keyID <= 0 {
		return RegenerateAPIKeyResult{}, fmt.Errorf("keyID must be positive")
	}

	payload := map[string]any{
		"key_id": keyID,
	}

	var out regenerateAPIKeyResponse
	if err := c.requestJSON(ctx, http.MethodPost, "keys", "regenerate", nil, payload, &out); err != nil {
		return RegenerateAPIKeyResult{}, err
	}
	if err := ensureSuccess("regenerate api key", out.baseResponse); err != nil {
		return RegenerateAPIKeyResult{}, err
	}

	return RegenerateAPIKeyResult{
		APIKey:    out.APIKey,
		APISecret: out.APISecret,
		Warning:   strings.TrimSpace(out.Warning),
		Message:   strings.TrimSpace(out.Message),
	}, nil
}
