package app

import (
	"context"
	"fmt"
	"strings"

	"dnsherene/pkg/dnshe"
	"dnsherene/pkg/notifier"
)

// Result 表示一次续期任务的结构化结果。
type Result struct {
	Matched     int
	Renewed     int
	Failed      int
	DryRun      bool
	RenewedList []notifier.RenewedDomain
	FailedList  []notifier.FailedDomain
}

// Service 负责续期任务编排：选目标、调用 SDK。
type Service struct {
	dnsClient *dnshe.Client
}

type renewFailuresError struct {
	Count   int
	Reasons []string
}

// Error 返回续期失败数量和摘要原因。
func (e *renewFailuresError) Error() string {
	if e == nil || e.Count <= 0 {
		return "renew requests failed"
	}

	message := fmt.Sprintf("%d renew request(s) failed", e.Count)
	if len(e.Reasons) > 0 {
		message += ": " + strings.Join(e.Reasons, ", ")
	}
	return message
}

// NewService 创建续期服务实例。
func NewService(dnsClient *dnshe.Client) (*Service, error) {
	if dnsClient == nil {
		return nil, fmt.Errorf("dns client is required")
	}
	return &Service{dnsClient: dnsClient}, nil
}

// Run 执行一次续期任务。
func (s *Service) Run(ctx context.Context, dryRun bool) (Result, error) {
	result := Result{}

	targets, err := s.dnsClient.ListSubdomains(ctx)
	if err != nil {
		return result, err
	}

	if len(targets) == 0 {
		return result, nil
	}

	result.Matched = len(targets)
	result.DryRun = dryRun

	if dryRun {
		return result, nil
	}

	result.RenewedList = make([]notifier.RenewedDomain, 0, len(targets))
	result.FailedList = make([]notifier.FailedDomain, 0)
	failureReasons := make([]string, 0, 3)
	failureReasonSeen := make(map[string]struct{})

	for _, target := range targets {
		renewResult, renewErr := s.dnsClient.RenewSubdomain(ctx, target.ID)
		if renewErr != nil {
			result.Failed++
			domain := target.DomainName()
			if domain == "" {
				domain = fmt.Sprintf("id-%d", target.ID)
			}
			result.FailedList = append(result.FailedList, notifier.FailedDomain{
				Domain: domain,
				Reason: renewErr.Error(),
			})

			reason := strings.TrimSpace(renewErr.Error())
			if reason == "" {
				reason = "request failed"
			}
			if _, ok := failureReasonSeen[reason]; !ok && len(failureReasons) < 3 {
				failureReasonSeen[reason] = struct{}{}
				failureReasons = append(failureReasons, reason)
			}
			continue
		}

		result.Renewed++
		successDomain := strings.TrimSpace(renewResult.Subdomain)
		if successDomain == "" {
			successDomain = target.DomainName()
		}
		result.RenewedList = append(result.RenewedList, notifier.RenewedDomain{
			Domain:        successDomain,
			NewExpiresAt:  renewResult.NewExpiresAt,
			RemainingDays: renewResult.RemainingDays,
		})
	}

	if result.Failed > 0 {
		return result, &renewFailuresError{
			Count:   result.Failed,
			Reasons: failureReasons,
		}
	}
	return result, nil
}
