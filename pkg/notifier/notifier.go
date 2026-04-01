package notifier

import (
	"context"

	"dnsherene/internal/report"
)

// Notifier 抽象统一通知能力，业务层只传结构化结果。
type Notifier interface {
	Notify(ctx context.Context, info report.Info) error
}
