package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"dnsherene/internal/app"
	"dnsherene/internal/config"
	"dnsherene/pkg/dnshe"
	"dnsherene/pkg/notifier"
)

type publicSummary struct {
	Renewed int
}

// run 顺序执行全部 API 凭证的续期流程。
//
// 当存在多组凭证时，单组失败不会中断后续组执行，最终统一聚合错误返回。
func run(cfg config.Config) (publicSummary, error) {
	summary := publicSummary{}
	ctx := context.Background()

	var runErrs []error
	total := len(cfg.Credentials)
	accounts := make([]notifier.AccountInfo, 0, total)
	for i, cred := range cfg.Credentials {
		account, err := runAccount(ctx, cfg, cred, i+1, total)
		summary.Renewed += account.Renewed
		accounts = append(accounts, account)
		if err != nil {
			runErrs = append(runErrs, err)
		}
	}

	info := notifier.Info{
		RenewedTotal: summary.Renewed,
		Accounts:     accounts,
	}
	for _, n := range notifier.Builtins {
		_ = n.Notify(ctx, cfg, info)
	}

	if len(runErrs) > 0 {
		return summary, errors.Join(runErrs...)
	}
	return summary, nil
}

// runAccount 执行单个 API 账号的一次完整续期任务。
func runAccount(
	ctx context.Context,
	cfg config.Config,
	cred config.APICredential,
	index int,
	total int,
) (notifier.AccountInfo, error) {
	account := notifier.AccountInfo{
		Index:        index,
		Total:        total,
		APIKeyMasked: maskAPIKey(cred.APIKey),
	}

	dnsClient, err := dnshe.NewClient(dnshe.Config{
		BaseURL:   cfg.APIBaseURL,
		APIKey:    cred.APIKey,
		APISecret: cred.APISecret,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	})
	if err != nil {
		account.Error = err.Error()
		return account, fmt.Errorf("init dnshe sdk for api[%d] failed: %w", index, err)
	}

	service, err := app.NewService(dnsClient)
	if err != nil {
		account.Error = err.Error()
		return account, fmt.Errorf("init service for api[%d] failed: %w", index, err)
	}

	result, err := service.Run(ctx, cfg.DryRun)
	account.Matched = result.Matched
	account.Renewed = result.Renewed
	account.Failed = result.Failed
	account.DryRun = result.DryRun
	account.RenewedList = result.RenewedList
	account.FailedList = result.FailedList
	if err != nil {
		account.Error = err.Error()
		return account, fmt.Errorf("run renew for api[%d] failed: %w", index, err)
	}

	return account, nil
}
