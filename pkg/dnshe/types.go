package dnshe

import (
	"strconv"
	"strings"
)

// Subdomain 表示 DNSHE 返回的子域名对象。
type Subdomain struct {
	ID         int    `json:"id"`
	Subdomain  string `json:"subdomain"`
	Rootdomain string `json:"rootdomain"`
	FullDomain string `json:"full_domain"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// DomainName 返回子域名的可展示完整域名。
func (s Subdomain) DomainName() string {
	if full := strings.TrimSpace(s.FullDomain); full != "" {
		return full
	}
	if sub := strings.TrimSpace(s.Subdomain); sub != "" {
		if root := strings.TrimSpace(s.Rootdomain); root != "" {
			return sub + "." + root
		}
		return sub
	}
	if root := strings.TrimSpace(s.Rootdomain); root != "" {
		return root
	}
	return ""
}

// SubdomainDetail 表示子域名详情以及其 DNS 记录集合。
type SubdomainDetail struct {
	Subdomain  Subdomain
	DNSRecords []DNSRecord
	DNSCount   int
}

// RegisterSubdomainRequest 定义注册子域名请求参数。
type RegisterSubdomainRequest struct {
	Subdomain  string
	Rootdomain string
}

// RegisterSubdomainResult 定义注册子域名返回结果。
type RegisterSubdomainResult struct {
	SubdomainID int
	FullDomain  string
	Message     string
}

// RenewResult 定义续期子域名返回结果。
type RenewResult struct {
	SubdomainID       int
	Message           string
	Subdomain         string
	PreviousExpiresAt string
	NewExpiresAt      string
	RenewedAt         string
	NeverExpires      bool
	Status            string
	RemainingDays     int
}

// DNSRecord 表示 DNS 记录对象。
type DNSRecord struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	TTL       int    `json:"ttl"`
	Priority  *int   `json:"priority"`
	Proxied   bool   `json:"proxied"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateDNSRecordRequest 定义创建 DNS 记录请求参数。
type CreateDNSRecordRequest struct {
	SubdomainID int
	Type        string
	Content     string
	Name        string
	TTL         int
	Priority    *int
}

// CreateDNSRecordResult 定义创建 DNS 记录返回结果。
type CreateDNSRecordResult struct {
	RecordID int
	Message  string
}

// UpdateDNSRecordRequest 定义更新 DNS 记录请求参数。
type UpdateDNSRecordRequest struct {
	RecordID int
	Content  *string
	TTL      *int
	Priority *int
}

// APIKey 表示 API Key 元数据对象。
type APIKey struct {
	ID           int    `json:"id"`
	KeyName      string `json:"key_name"`
	APIKey       string `json:"api_key"`
	Status       string `json:"status"`
	RequestCount int    `json:"request_count"`
	LastUsedAt   string `json:"last_used_at"`
	CreatedAt    string `json:"created_at"`
}

// CreateAPIKeyRequest 定义创建 API Key 请求参数。
type CreateAPIKeyRequest struct {
	KeyName     string
	IPWhitelist string
}

// CreateAPIKeyResult 定义创建 API Key 返回结果。
type CreateAPIKeyResult struct {
	APIKey    string
	APISecret string
	Warning   string
	Message   string
}

// RegenerateAPIKeyResult 定义重置 API Key 返回结果。
type RegenerateAPIKeyResult struct {
	APIKey    string
	APISecret string
	Warning   string
	Message   string
}

// Quota 表示免费域名额度信息。
type Quota struct {
	Used        int `json:"used"`
	Base        int `json:"base"`
	InviteBonus int `json:"invite_bonus"`
	Total       int `json:"total"`
	Available   int `json:"available"`
}

// APIError 表示 DNSHE API 返回的结构化错误。
//
// 当服务返回非 2xx HTTP 状态码，或者返回 success=false 的业务错误时，
// SDK 会优先返回该类型，便于调用方通过 errors.As 读取限流等上下文字段。
type APIError struct {
	Operation string

	StatusCode int
	ErrorText  string
	Message    string
	Limit      *int
	Remaining  *int
	ResetAt    string
	RawBody    string
}

// Error 返回适合日志和上层处理的错误字符串。
func (e *APIError) Error() string {
	if e == nil {
		return "unknown error"
	}
	msg := firstNonEmpty(e.ErrorText, e.Message, e.RawBody, "unknown error")

	if e.Operation != "" {
		return e.Operation + " failed: " + msg
	}
	if e.StatusCode > 0 {
		return "http " + strconv.Itoa(e.StatusCode) + ": " + msg
	}
	return msg
}
