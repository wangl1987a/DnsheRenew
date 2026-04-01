package dnshe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type listSubdomainsResponse struct {
	baseResponse
	Count      int         `json:"count"`
	Subdomains []Subdomain `json:"subdomains"`
}

type registerSubdomainResponse struct {
	baseResponse
	SubdomainID int    `json:"subdomain_id"`
	FullDomain  string `json:"full_domain"`
}

type getSubdomainResponse struct {
	baseResponse
	Subdomain  Subdomain   `json:"subdomain"`
	DNSRecords []DNSRecord `json:"dns_records"`
	DNSCount   int         `json:"dns_count"`
}

type renewSubdomainResponse struct {
	baseResponse
	SubdomainID       int    `json:"subdomain_id"`
	Subdomain         string `json:"subdomain"`
	PreviousExpiresAt string `json:"previous_expires_at"`
	NewExpiresAt      string `json:"new_expires_at"`
	RemainingDays     int    `json:"remaining_days"`
}

// ListSubdomains 获取当前账号可见的子域名列表。
func (c *Client) ListSubdomains(ctx context.Context) ([]Subdomain, error) {
	var out listSubdomainsResponse
	if err := c.requestJSON(ctx, http.MethodGet, "subdomains", "list", nil, nil, &out); err != nil {
		return nil, err
	}
	if err := ensureSuccess("list subdomains", out.baseResponse); err != nil {
		return nil, err
	}
	return out.Subdomains, nil
}

// RegisterSubdomain 注册新的子域名。
func (c *Client) RegisterSubdomain(ctx context.Context, req RegisterSubdomainRequest) (RegisterSubdomainResult, error) {
	subdomain := strings.TrimSpace(req.Subdomain)
	if subdomain == "" {
		return RegisterSubdomainResult{}, fmt.Errorf("subdomain is required")
	}
	rootdomain := strings.TrimSpace(req.Rootdomain)
	if rootdomain == "" {
		return RegisterSubdomainResult{}, fmt.Errorf("rootdomain is required")
	}

	payload := map[string]any{
		"subdomain":  subdomain,
		"rootdomain": rootdomain,
	}

	var out registerSubdomainResponse
	if err := c.requestJSON(ctx, http.MethodPost, "subdomains", "register", nil, payload, &out); err != nil {
		return RegisterSubdomainResult{}, err
	}
	if err := ensureSuccess("register subdomain", out.baseResponse); err != nil {
		return RegisterSubdomainResult{}, err
	}

	return RegisterSubdomainResult{
		SubdomainID: out.SubdomainID,
		FullDomain:  out.FullDomain,
		Message:     strings.TrimSpace(out.Message),
	}, nil
}

// GetSubdomain 获取单个子域名详情及其 DNS 记录。
func (c *Client) GetSubdomain(ctx context.Context, subdomainID int) (SubdomainDetail, error) {
	if subdomainID <= 0 {
		return SubdomainDetail{}, fmt.Errorf("subdomainID must be positive")
	}

	query := url.Values{}
	query.Set("subdomain_id", strconv.Itoa(subdomainID))

	var out getSubdomainResponse
	if err := c.requestJSON(ctx, http.MethodGet, "subdomains", "get", query, nil, &out); err != nil {
		return SubdomainDetail{}, err
	}
	if err := ensureSuccess("get subdomain", out.baseResponse); err != nil {
		return SubdomainDetail{}, err
	}

	return SubdomainDetail{
		Subdomain:  out.Subdomain,
		DNSRecords: out.DNSRecords,
		DNSCount:   out.DNSCount,
	}, nil
}

// DeleteSubdomain 删除指定子域名。
func (c *Client) DeleteSubdomain(ctx context.Context, subdomainID int) error {
	if subdomainID <= 0 {
		return fmt.Errorf("subdomainID must be positive")
	}

	payload := map[string]any{
		"subdomain_id": subdomainID,
	}

	var out baseResponse
	if err := c.requestJSON(ctx, http.MethodPost, "subdomains", "delete", nil, payload, &out); err != nil {
		return err
	}
	return ensureSuccess("delete subdomain", out)
}

// RenewSubdomain 按子域名 ID 发起续期请求。
func (c *Client) RenewSubdomain(ctx context.Context, subdomainID int) (RenewResult, error) {
	if subdomainID <= 0 {
		return RenewResult{}, fmt.Errorf("subdomainID must be positive")
	}

	payload := map[string]any{
		"subdomain_id": subdomainID,
	}

	var out renewSubdomainResponse
	if err := c.requestJSON(ctx, http.MethodPost, "subdomains", "renew", nil, payload, &out); err != nil {
		return RenewResult{}, err
	}
	if err := ensureSuccess("renew subdomain", out.baseResponse); err != nil {
		return RenewResult{}, err
	}

	return RenewResult{
		SubdomainID:       out.SubdomainID,
		Subdomain:         out.Subdomain,
		PreviousExpiresAt: out.PreviousExpiresAt,
		NewExpiresAt:      out.NewExpiresAt,
		RemainingDays:     out.RemainingDays,
	}, nil
}
