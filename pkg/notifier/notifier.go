package notifier

import (
	"context"
	"time"

	"dnsherene/internal/config"
)

// Info 是一次完整续期任务的通知载荷。
//
// 调用方只负责填充结构化结果，具体输出格式由各个通知实现决定。
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
	RenewedList  []RenewedDomain `json:"renewed_list,omitempty"`
	FailedList   []FailedDomain  `json:"failed_list,omitempty"`
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

// Notifier 抽象统一通知能力，业务层只传结构化结果。
type Notifier interface {
	Notify(ctx context.Context, cfg config.Config, info Info) error
}
