package notification

import (
	"context"
	"errors"
	"fmt"

	"dnsherene/internal/config"
	"dnsherene/internal/report"
)

// Notifier 抽象统一通知能力。
type Notifier interface {
	Notify(ctx context.Context, info report.Info) error
}

// Manager 负责根据配置分发通知。
type Manager struct {
	notifiers []Notifier
}

// NewManager 根据通知配置创建通知管理器。
func NewManager(cfg config.NotificationConfig) (Manager, error) {
	notifiers := make([]Notifier, 0, 3)

	if cfg.Console.Enabled {
		notifiers = append(notifiers, Console{enabled: true})
	}
	if cfg.Telegram != nil {
		notifiers = append(notifiers, Telegram{config: *cfg.Telegram})
	}
	if cfg.Webhook != nil {
		notifiers = append(notifiers, Webhook{config: *cfg.Webhook})
	}

	if cfg.Mail != nil {
		return Manager{}, fmt.Errorf("mail notification is not implemented yet")
	}
	if cfg.Lark != nil {
		return Manager{}, fmt.Errorf("lark notification is not implemented yet")
	}

	return Manager{notifiers: notifiers}, nil
}

// Notify 将结构化报告发送给全部已配置通知器。
func (m Manager) Notify(ctx context.Context, info report.Info) error {
	var notifyErrs []error
	for _, n := range m.notifiers {
		if err := n.Notify(ctx, info); err != nil {
			notifyErrs = append(notifyErrs, err)
		}
	}
	return errors.Join(notifyErrs...)
}
