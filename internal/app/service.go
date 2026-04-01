package app

import (
	"context"
	"fmt"
	"strings"

	"dnsherene/pkg/dnshe"
	"dnsherene/pkg/notifier"
)

// Config 定义一次续期任务的执行参数。
type Config struct {
	// IDListRaw 指定优先续期的子域名 ID 列表（逗号分隔）。
	// 该字段非空时，将忽略 RootdomainFilter/SubdomainFilter。
	IDListRaw string
	// RootdomainFilter 是根域名过滤条件（仅在 IDListRaw 为空时生效）。
	RootdomainFilter string
	// SubdomainFilter 是子域名前缀过滤条件（仅在 IDListRaw 为空时生效）。
	SubdomainFilter string
	// DryRun 为 true 时只做匹配与通知，不执行实际续期请求。
	DryRun bool
}

// Result 表示一次续期任务的统计结果。
type Result struct {
	// Matched 是匹配到的目标子域名数量。
	Matched int
	// Renewed 是续期成功的数量。
	Renewed int
	// Failed 是续期失败的数量。
	Failed int
}

// Service 负责续期任务编排：选目标、调用 SDK、发送通知。
type Service struct {
	dnsClient *dnshe.Client
	notifier  notifier.Notifier
}

type renewDetail struct {
	Domain        string `json:"domain"`
	NewExpiresAt  string `json:"new_expires_at"`
	RemainingDays int    `json:"remaining_days"`
}

type renewFailureDetail struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

// NewService 创建续期服务实例。
func NewService(dnsClient *dnshe.Client, n notifier.Notifier) (*Service, error) {
	if dnsClient == nil {
		return nil, fmt.Errorf("dns client is required")
	}
	return &Service{
		dnsClient: dnsClient,
		notifier:  n,
	}, nil
}

// Run 执行一次续期任务。
//
// 执行流程：
// 1. 拉取子域名列表并按配置筛选目标。
// 2. dry-run 模式下仅返回统计结果，并发送一条详细通知。
// 3. 对目标逐个续期，并在结束后按每个 API 汇总详细通知。
func (s *Service) Run(ctx context.Context, cfg Config) (Result, error) {
	result := Result{}

	allSubdomains, err := s.dnsClient.ListSubdomains(ctx)
	if err != nil {
		s.notify(ctx, notifier.Event{
			Level:   notifier.LevelError,
			Title:   "Renew detail",
			Message: "list subdomains failed",
			Fields: map[string]any{
				"updated":     0,
				"not_updated": 0,
				"error":       err.Error(),
			},
		})
		return result, err
	}

	targets, err := selectTargets(allSubdomains, cfg.IDListRaw, cfg.RootdomainFilter, cfg.SubdomainFilter)
	if err != nil {
		s.notify(ctx, notifier.Event{
			Level:   notifier.LevelError,
			Title:   "Renew detail",
			Message: "select targets failed",
			Fields: map[string]any{
				"updated":     0,
				"not_updated": 0,
				"error":       err.Error(),
			},
		})
		return result, err
	}
	if len(targets) == 0 {
		err = fmt.Errorf("no subdomains matched selection")
		s.notify(ctx, notifier.Event{
			Level:   notifier.LevelError,
			Title:   "Renew detail",
			Message: err.Error(),
			Fields: map[string]any{
				"updated":     0,
				"not_updated": 0,
			},
		})
		return result, err
	}

	result.Matched = len(targets)

	if cfg.DryRun {
		s.notify(ctx, notifier.Event{
			Level:   notifier.LevelInfo,
			Title:   "Renew detail",
			Message: "dry run completed",
			Fields: map[string]any{
				"matched":      result.Matched,
				"updated":      0,
				"not_updated":  result.Matched,
				"dry_run":      true,
				"target_count": result.Matched,
			},
		})
		return result, nil
	}

	failures := make([]string, 0)
	updatedDetails := make([]renewDetail, 0, len(targets))
	notUpdatedDetails := make([]renewFailureDetail, 0)

	for _, target := range targets {
		renewResult, renewErr := s.dnsClient.RenewSubdomain(ctx, target.ID)
		if renewErr != nil {
			result.Failed++
			domain := target.DomainName()
			if domain == "" {
				domain = fmt.Sprintf("id-%d", target.ID)
			}
			failures = append(failures, fmt.Sprintf("id=%d domain=%s err=%v", target.ID, domain, renewErr))
			notUpdatedDetails = append(notUpdatedDetails, renewFailureDetail{
				Domain: domain,
				Reason: renewErr.Error(),
			})
			continue
		}

		result.Renewed++
		successDomain := strings.TrimSpace(renewResult.Subdomain)
		if successDomain == "" {
			successDomain = target.DomainName()
		}
		updatedDetails = append(updatedDetails, renewDetail{
			Domain:        successDomain,
			NewExpiresAt:  renewResult.NewExpiresAt,
			RemainingDays: renewResult.RemainingDays,
		})
	}

	level := notifier.LevelInfo
	message := "renew completed"
	if result.Failed > 0 {
		level = notifier.LevelError
		message = "renew completed with failures"
	}

	s.notify(ctx, notifier.Event{
		Level:   level,
		Title:   "Renew detail",
		Message: message,
		Fields: map[string]any{
			"matched":             result.Matched,
			"updated":             result.Renewed,
			"not_updated":         result.Failed,
			"updated_domains":     updatedDetails,
			"not_updated_domains": notUpdatedDetails,
		},
	})

	if result.Failed > 0 {
		errMsg := fmt.Sprintf("%d renew request(s) failed", result.Failed)
		if len(failures) > 0 {
			errMsg = errMsg + ": " + strings.Join(failures, "; ")
		}
		return result, fmt.Errorf(errMsg)
	}
	return result, nil
}

// notify 负责将详细事件发送给通知模块。
//
// 通知失败不会写入公共日志，避免在 GitHub Actions 日志中泄露敏感信息。
func (s *Service) notify(ctx context.Context, event notifier.Event) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Notify(ctx, event)
}
