package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	publicURLPattern    = regexp.MustCompile(`https?://\S+`)
	publicAPIKeyPattern = regexp.MustCompile(`\bcfsd_[A-Za-z0-9_-]+\b`)
	publicDomainPattern = regexp.MustCompile(`\b(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}\b`)
	publicSpacePattern  = regexp.MustCompile(`\s+`)
)

// maskAPIKey 返回用于日志展示的脱敏 API Key。
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

// writePublicErrorReport 将聚合错误按脱敏后的公开格式写入输出流。
func writePublicErrorReport(w io.Writer, err error) {
	items := splitPublicErrors(err)
	if len(items) == 0 {
		items = []error{err}
	}

	_, _ = fmt.Fprintf(w, "error_count=%d\n", len(items))
	for i, item := range items {
		_, _ = fmt.Fprintf(w, "error[%d]=%s\n", i+1, sanitizePublicError(item))
	}
}

// splitPublicErrors 递归展开 errors.Join 形成的错误列表。
func splitPublicErrors(err error) []error {
	if err == nil {
		return nil
	}

	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		result := make([]error, 0)
		for _, item := range multi.Unwrap() {
			result = append(result, splitPublicErrors(item)...)
		}
		return result
	}

	return []error{err}
}

// sanitizePublicError 对公开错误信息做 URL、域名和 API Key 脱敏。
func sanitizePublicError(err error) string {
	if err == nil {
		return "unknown error"
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "unknown error"
	}

	message = publicURLPattern.ReplaceAllString(message, "<redacted_url>")
	message = publicAPIKeyPattern.ReplaceAllStringFunc(message, func(value string) string {
		return maskAPIKey(value)
	})
	message = publicDomainPattern.ReplaceAllString(message, "<redacted_domain>")
	message = publicSpacePattern.ReplaceAllString(message, " ")
	message = strings.TrimSpace(message)

	if message == "" {
		return "unknown error"
	}
	if len(message) > 240 {
		return message[:240] + "..."
	}
	return message
}
