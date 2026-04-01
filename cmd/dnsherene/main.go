package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"dnsherene/internal/app"
	"dnsherene/internal/config"
	"dnsherene/pkg/dnshe"
	"dnsherene/pkg/notifier"
)

type publicSummary struct {
	Renewed int
}

func main() {
	summary, err := run()
	fmt.Printf("renewed_total=%d\n", summary.Renewed)
	if err != nil {
		os.Exit(1)
	}
}

// run 负责组装依赖并顺序执行全部 API 凭证的续期流程。
//
// 当存在多组凭证时，单组失败不会中断后续组执行，最终统一聚合错误返回。
func run() (publicSummary, error) {
	summary := publicSummary{}

	cfg, err := config.Load()
	if err != nil {
		return summary, err
	}

	n, err := buildNotifier(cfg)
	if err != nil {
		return summary, fmt.Errorf("init notifier failed: %w", err)
	}

	var runErrs []error
	total := len(cfg.Credentials)
	for i, cred := range cfg.Credentials {
		dnsClient, clientErr := dnshe.NewClient(dnshe.Config{
			BaseURL:   cfg.APIBaseURL,
			APIKey:    cred.APIKey,
			APISecret: cred.APISecret,
			HTTPClient: &http.Client{
				Timeout: 20 * time.Second,
			},
		})
		if clientErr != nil {
			runErrs = append(runErrs, fmt.Errorf("init dnshe sdk for api[%d] failed: %w", i+1, clientErr))
			continue
		}

		scopedNotifier := newIndexedNotifier(n, i+1, total, cred.APIKey)
		service, serviceErr := app.NewService(dnsClient, scopedNotifier)
		if serviceErr != nil {
			runErrs = append(runErrs, fmt.Errorf("init service for api[%d] failed: %w", i+1, serviceErr))
			continue
		}

		result, serviceErr := service.Run(context.Background(), app.Config{
			IDListRaw:        cfg.SubdomainIDs,
			RootdomainFilter: cfg.RootdomainFilter,
			SubdomainFilter:  cfg.SubdomainFilter,
			DryRun:           cfg.DryRun,
		})
		summary.Renewed += result.Renewed
		if serviceErr != nil {
			runErrs = append(runErrs, fmt.Errorf("run renew for api[%d] failed: %w", i+1, serviceErr))
		}
	}

	if len(runErrs) > 0 {
		return summary, errors.Join(runErrs...)
	}
	return summary, nil
}

// buildNotifier 根据配置构建详细通知链路。
//
// 公共日志由 main 单独输出，通知模块只负责私有明细。
func buildNotifier(cfg config.Config) (notifier.Notifier, error) {
	if cfg.NotifyWebhookURL == "" {
		return nil, nil
	}

	webhook, err := notifier.NewWebhook(cfg.NotifyWebhookURL, cfg.NotifyWebhookToken, &http.Client{
		Timeout: 15 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

// indexedNotifier 为详细通知补充 API 维度信息，便于区分多组凭证。
type indexedNotifier struct {
	base      notifier.Notifier
	index     int
	total     int
	maskedKey string
}

// newIndexedNotifier 创建带 API 标记的通知器包装。
func newIndexedNotifier(base notifier.Notifier, index int, total int, apiKey string) notifier.Notifier {
	if base == nil {
		return nil
	}
	return &indexedNotifier{
		base:      base,
		index:     index,
		total:     total,
		maskedKey: maskAPIKey(apiKey),
	}
}

// Notify 在详细通知中附加 API 下标和脱敏后的凭证标识。
func (i *indexedNotifier) Notify(ctx context.Context, event notifier.Event) error {
	fields := make(map[string]any, len(event.Fields)+3)
	for k, v := range event.Fields {
		fields[k] = v
	}
	fields["api_index"] = i.index
	fields["api_total"] = i.total
	fields["api_key_masked"] = i.maskedKey

	event.Fields = fields
	return i.base.Notify(ctx, event)
}

func maskAPIKey(apiKey string) string {
	switch {
	case len(apiKey) <= 4:
		return "***"
	case len(apiKey) <= 8:
		return apiKey[:2] + "***" + apiKey[len(apiKey)-2:]
	default:
		return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
	}
}
