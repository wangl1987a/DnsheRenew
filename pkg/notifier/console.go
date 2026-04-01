package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"dnsherene/internal/config"
)

type Console struct {
	outWriter io.Writer
	errWriter io.Writer
}

// Notify 在调试模式下把结构化续期结果打印到控制台。
func (c Console) Notify(_ context.Context, cfg config.Config, info Info) error {
	if !cfg.Debug {
		return nil
	}

	ts := info.GeneratedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	writer := c.outWriter
	if writer == nil {
		writer = os.Stdout
	}
	if hasAccountErrors(info) {
		writer = c.errWriter
		if writer == nil {
			writer = os.Stderr
		}
	}

	header := fmt.Sprintf(
		"[%s] renew summary: renewed_total=%d accounts=%d",
		ts.Format(time.RFC3339),
		info.RenewedTotal,
		len(info.Accounts),
	)
	if _, err := fmt.Fprintln(writer, header); err != nil {
		return err
	}

	for _, account := range info.Accounts {
		line := fmt.Sprintf(
			"api[%d/%d] key=%s matched=%d renewed=%d failed=%d",
			account.Index,
			account.Total,
			fallback(account.APIKeyMasked, "***"),
			account.Matched,
			account.Renewed,
			account.Failed,
		)
		if account.DryRun {
			line += " dry_run=true"
		}
		if account.Error != "" {
			line += " error=" + account.Error
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}

		if len(account.RenewedList) > 0 {
			if err := writeJSONLine(writer, "renewed_list", account.RenewedList); err != nil {
				return err
			}
		}
		if len(account.FailedList) > 0 {
			if err := writeJSONLine(writer, "failed_list", account.FailedList); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeJSONLine 将结构化字段编码为单行 JSON 输出。
func writeJSONLine(w io.Writer, label string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s failed: %w", label, err)
	}
	_, err = fmt.Fprintf(w, "%s=%s\n", label, raw)
	return err
}

// hasAccountErrors 判断通知内容中是否包含失败账号或失败域名。
func hasAccountErrors(info Info) bool {
	for _, account := range info.Accounts {
		if strings.TrimSpace(account.Error) != "" || account.Failed > 0 {
			return true
		}
	}
	return false
}

// fallback 在输入为空时返回兜底值。
func fallback(value string, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
