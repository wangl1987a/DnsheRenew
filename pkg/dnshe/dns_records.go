package dnshe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var supportedDNSRecordTypes = map[string]struct{}{
	"A":     {},
	"AAAA":  {},
	"CNAME": {},
	"MX":    {},
	"TXT":   {},
}

type listDNSRecordsResponse struct {
	baseResponse
	Count   int         `json:"count"`
	Records []DNSRecord `json:"records"`
}

type createDNSRecordResponse struct {
	baseResponse
	RecordID int `json:"record_id"`
}

// ListDNSRecords 获取指定子域名下的 DNS 记录列表。
func (c *Client) ListDNSRecords(ctx context.Context, subdomainID int) ([]DNSRecord, error) {
	if subdomainID <= 0 {
		return nil, fmt.Errorf("subdomainID must be positive")
	}

	query := url.Values{}
	query.Set("subdomain_id", strconv.Itoa(subdomainID))

	var out listDNSRecordsResponse
	if err := c.requestJSON(ctx, http.MethodGet, "dns_records", "list", query, nil, &out); err != nil {
		return nil, err
	}
	if err := ensureSuccess("list dns records", out.baseResponse); err != nil {
		return nil, err
	}
	return out.Records, nil
}

// CreateDNSRecord 为子域名创建 DNS 记录。
func (c *Client) CreateDNSRecord(ctx context.Context, req CreateDNSRecordRequest) (CreateDNSRecordResult, error) {
	if req.SubdomainID <= 0 {
		return CreateDNSRecordResult{}, fmt.Errorf("subdomainID must be positive")
	}

	recordType, err := normalizeRecordType(req.Type)
	if err != nil {
		return CreateDNSRecordResult{}, err
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return CreateDNSRecordResult{}, fmt.Errorf("content is required")
	}
	if req.TTL < 0 {
		return CreateDNSRecordResult{}, fmt.Errorf("ttl must be positive")
	}
	if req.Priority != nil && *req.Priority < 0 {
		return CreateDNSRecordResult{}, fmt.Errorf("priority must be zero or greater")
	}
	if recordType == "MX" && req.Priority == nil {
		return CreateDNSRecordResult{}, fmt.Errorf("priority is required for MX records")
	}

	payload := map[string]any{
		"subdomain_id": req.SubdomainID,
		"type":         recordType,
		"content":      content,
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		payload["name"] = name
	}
	if req.TTL > 0 {
		payload["ttl"] = req.TTL
	}
	if req.Priority != nil {
		payload["priority"] = *req.Priority
	}

	var out createDNSRecordResponse
	if err := c.requestJSON(ctx, http.MethodPost, "dns_records", "create", nil, payload, &out); err != nil {
		return CreateDNSRecordResult{}, err
	}
	if err := ensureSuccess("create dns record", out.baseResponse); err != nil {
		return CreateDNSRecordResult{}, err
	}

	return CreateDNSRecordResult{
		RecordID: out.RecordID,
		Message:  strings.TrimSpace(out.Message),
	}, nil
}

// UpdateDNSRecord 更新 DNS 记录字段。
func (c *Client) UpdateDNSRecord(ctx context.Context, req UpdateDNSRecordRequest) error {
	if req.RecordID <= 0 {
		return fmt.Errorf("recordID must be positive")
	}

	payload := map[string]any{
		"record_id": req.RecordID,
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content == "" {
			return fmt.Errorf("content must not be empty")
		}
		payload["content"] = content
	}
	if req.TTL != nil {
		if *req.TTL <= 0 {
			return fmt.Errorf("ttl must be positive")
		}
		payload["ttl"] = *req.TTL
	}
	if req.Priority != nil {
		if *req.Priority < 0 {
			return fmt.Errorf("priority must be zero or greater")
		}
		payload["priority"] = *req.Priority
	}
	if len(payload) == 1 {
		return fmt.Errorf("at least one field must be set for update")
	}

	var out baseResponse
	if err := c.requestJSON(ctx, http.MethodPost, "dns_records", "update", nil, payload, &out); err != nil {
		return err
	}
	return ensureSuccess("update dns record", out)
}

// normalizeRecordType 规范化并校验 DNS 记录类型。
func normalizeRecordType(raw string) (string, error) {
	recordType := strings.ToUpper(strings.TrimSpace(raw))
	if recordType == "" {
		return "", fmt.Errorf("type is required")
	}
	if _, ok := supportedDNSRecordTypes[recordType]; !ok {
		return "", fmt.Errorf("unsupported dns record type: %s", recordType)
	}
	return recordType, nil
}

// DeleteDNSRecord 删除指定 DNS 记录。
func (c *Client) DeleteDNSRecord(ctx context.Context, recordID int) error {
	if recordID <= 0 {
		return fmt.Errorf("recordID must be positive")
	}

	payload := map[string]any{
		"record_id": recordID,
	}

	var out baseResponse
	if err := c.requestJSON(ctx, http.MethodPost, "dns_records", "delete", nil, payload, &out); err != nil {
		return err
	}
	return ensureSuccess("delete dns record", out)
}
