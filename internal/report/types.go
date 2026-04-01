package report

import "time"

// Info 是一次完整续期任务的结构化报告。
type Info struct {
	GeneratedAt  time.Time     `json:"generated_at"`
	RenewedTotal int           `json:"renewed_total"`
	Accounts     []AccountInfo `json:"accounts,omitempty"`
}

// AccountInfo 表示单组 API 凭证的续期结果。
type AccountInfo struct {
	Index        int             `json:"index"`
	Total        int             `json:"total"`
	APIKeyMasked string          `json:"api_key_masked,omitempty"`
	Matched      int             `json:"matched"`
	Renewed      int             `json:"renewed"`
	Failed       int             `json:"failed"`
	DryRun       bool            `json:"dry_run,omitempty"`
	Error        string          `json:"error,omitempty"`
	Domains      []DomainInfo    `json:"domains,omitempty"`
	RenewedList  []RenewedDomain `json:"renewed_list,omitempty"`
	FailedList   []FailedDomain  `json:"failed_list,omitempty"`
}

// DomainInfo 表示账号下子域名的当前到期信息。
type DomainInfo struct {
	Domain        string `json:"domain"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	RemainingDays *int   `json:"remaining_days,omitempty"`
}

// RenewedDomain 表示续期成功的域名信息。
type RenewedDomain struct {
	Domain        string `json:"domain"`
	NewExpiresAt  string `json:"new_expires_at,omitempty"`
	RemainingDays int    `json:"remaining_days,omitempty"`
}

// FailedDomain 表示续期失败的域名信息。
type FailedDomain struct {
	Domain string `json:"domain"`
	Reason string `json:"reason,omitempty"`
}
