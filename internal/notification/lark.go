package notification

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dnsherene/internal/config"
	"dnsherene/internal/report"

	notifylark "github.com/nikoksr/notify/service/lark"
)

type larkSender interface {
	Send(ctx context.Context, subject, message string) error
}

type Lark struct {
	config config.LarkConfig
	sender larkSender
}

// Notify 在配置了 Lark 时发送纯文本摘要通知。
func (l Lark) Notify(ctx context.Context, info report.Info) error {
	sender, err := l.resolveSender()
	if err != nil {
		return err
	}

	subject, message := buildLarkNotification(info)
	return sender.Send(ctx, subject, message)
}

func (l Lark) resolveSender() (larkSender, error) {
	if l.sender != nil {
		return l.sender, nil
	}

	switch l.config.Mode {
	case config.LarkModeWebhook:
		if strings.TrimSpace(l.config.WebhookURL) == "" {
			return nil, fmt.Errorf("lark webhook url is required")
		}
		return notifylark.NewWebhookService(strings.TrimSpace(l.config.WebhookURL)), nil
	case config.LarkModeCustomApp:
		if strings.TrimSpace(l.config.AppID) == "" || strings.TrimSpace(l.config.AppSecret) == "" {
			return nil, fmt.Errorf("lark custom app id and secret are required")
		}

		service := notifylark.NewCustomAppService(
			strings.TrimSpace(l.config.AppID),
			strings.TrimSpace(l.config.AppSecret),
		)

		receivers, err := buildLarkReceivers(l.config.ReceiverType, l.config.Receivers)
		if err != nil {
			return nil, err
		}
		service.AddReceivers(receivers...)
		return service, nil
	default:
		return nil, fmt.Errorf("unsupported lark notification mode: %s", l.config.Mode)
	}
}

func buildLarkReceivers(receiverType string, values []string) ([]*notifylark.ReceiverID, error) {
	values = filterNonEmpty(values)
	if len(values) == 0 {
		return nil, fmt.Errorf("lark receivers are required")
	}

	receivers := make([]*notifylark.ReceiverID, 0, len(values))
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(receiverType)) {
		case "open_id":
			receivers = append(receivers, notifylark.OpenID(value))
		case "user_id":
			receivers = append(receivers, notifylark.UserID(value))
		case "union_id":
			receivers = append(receivers, notifylark.UnionID(value))
		case "email":
			receivers = append(receivers, notifylark.Email(value))
		case "chat_id":
			receivers = append(receivers, notifylark.ChatID(value))
		default:
			return nil, fmt.Errorf("unsupported lark receiver type: %s", receiverType)
		}
	}

	return receivers, nil
}

func buildLarkNotification(info report.Info) (string, string) {
	return "DNSHE Renew Summary", buildLarkBody(info)
}

func buildLarkBody(info report.Info) string {
	ts := info.GeneratedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	failedAccounts := 0
	for _, account := range info.Accounts {
		if account.Failed > 0 || strings.TrimSpace(account.Error) != "" {
			failedAccounts++
		}
	}

	lines := []string{
		"Generated: " + ts.UTC().Format("2006-01-02 15:04:05 UTC"),
		fmt.Sprintf("Renewed total: %d", info.RenewedTotal),
		fmt.Sprintf("Accounts: %d", len(info.Accounts)),
		fmt.Sprintf("Accounts with issues: %d", failedAccounts),
	}

	for _, account := range info.Accounts {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("API [%d/%d] %s", account.Index, account.Total, fallback(account.APIKeyMasked, "***")))
		lines = append(lines, fmt.Sprintf("Matched: %d", account.Matched))
		lines = append(lines, fmt.Sprintf("Renewed: %d", account.Renewed))
		lines = append(lines, fmt.Sprintf("Failed: %d", account.Failed))
		if account.DryRun {
			lines = append(lines, "Dry run: true")
		}
		if strings.TrimSpace(account.Error) != "" {
			lines = append(lines, "Error: "+strings.TrimSpace(account.Error))
		}

		if len(account.Domains) > 0 {
			lines = append(lines, "Domains:")
			for _, domain := range account.Domains {
				meta := make([]string, 0, 2)
				if strings.TrimSpace(domain.ExpiresAt) != "" {
					meta = append(meta, "expires "+strings.TrimSpace(domain.ExpiresAt))
				}
				if domain.RemainingDays != nil {
					meta = append(meta, strconv.Itoa(*domain.RemainingDays)+"d")
				}
				if len(meta) > 0 {
					lines = append(lines, "- "+domain.Domain+" | "+strings.Join(meta, " | "))
					continue
				}
				lines = append(lines, "- "+domain.Domain)
			}
		}

		if len(account.RenewedList) > 0 {
			lines = append(lines, "Renewed:")
			for _, item := range account.RenewedList {
				line := "- " + item.Domain
				if strings.TrimSpace(item.NewExpiresAt) != "" {
					line += " -> " + strings.TrimSpace(item.NewExpiresAt)
				}
				if item.RemainingDays > 0 {
					line += " (" + strconv.Itoa(item.RemainingDays) + "d)"
				}
				lines = append(lines, line)
			}
		}

		if len(account.FailedList) > 0 {
			lines = append(lines, "Failed:")
			for _, item := range account.FailedList {
				line := "- " + item.Domain
				if strings.TrimSpace(item.Reason) != "" {
					line += " | " + strings.TrimSpace(item.Reason)
				}
				lines = append(lines, line)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func filterNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
