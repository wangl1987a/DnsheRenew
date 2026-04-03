package notification

import (
	"fmt"
	"strings"

	"dnsherene/internal/report"
)

// Request 是通知模块内部的统一请求模型。
//
// 业务层只需要产出 report.Info；通知层接收该结构后再完成各渠道的格式化。
// 这样可以避免把 subject/plain/html 一类的渠道格式要求泄漏到业务代码中。
type Request struct {
	Report report.Info
}

// Validate 校验通知请求的最小合法性。
func (r Request) Validate() error {
	if r.Report.GeneratedAt.IsZero() {
		return fmt.Errorf("notification report generated_at is required")
	}
	return nil
}

// AccountCount 返回本次通知包含的账号数。
func (r Request) AccountCount() int {
	return len(r.Report.Accounts)
}

// HasFailures 返回通知内容中是否存在账号级或域名级失败。
func (r Request) HasFailures() bool {
	for _, account := range r.Report.Accounts {
		if strings.TrimSpace(account.Error) != "" || account.Failed > 0 {
			return true
		}
	}
	return false
}
