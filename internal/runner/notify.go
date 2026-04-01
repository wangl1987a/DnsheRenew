package runner

import (
	"context"
	"errors"

	"dnsherene/internal/report"
	"dnsherene/pkg/notifier"
)

// Notify 将结构化报告发送给全部内建通知器。
func Notify(ctx context.Context, info report.Info) error {
	var notifyErrs []error
	for _, n := range notifier.Builtins {
		if err := n.Notify(ctx, info); err != nil {
			notifyErrs = append(notifyErrs, err)
		}
	}
	return errors.Join(notifyErrs...)
}
