package output

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

// MaskAPIKey 返回用于日志展示的脱敏 API Key。
func MaskAPIKey(apiKey string) string {
	switch {
	case len(apiKey) <= 4:
		return "***"
	case len(apiKey) <= 8:
		return apiKey[:2] + "***" + apiKey[len(apiKey)-2:]
	default:
		return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
	}
}

// WritePublicErrorReport 将聚合错误按脱敏后的公开格式写入输出流。
func WritePublicErrorReport(w io.Writer, err error) {
	WritePrefixedPublicErrorReport(w, "error", err)
}

// WritePrefixedPublicErrorReport 将聚合错误按带前缀的公开格式写入输出流。
func WritePrefixedPublicErrorReport(w io.Writer, prefix string, err error) {
	items := splitPublicErrors(err)
	if len(items) == 0 {
		items = []error{err}
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "error"
	}

	_, _ = fmt.Fprintf(w, "%s_count=%d\n", prefix, len(items))
	for i, item := range items {
		_, _ = fmt.Fprintf(w, "%s[%d]=%s\n", prefix, i+1, SanitizePublicError(item))
	}
}

// SanitizePublicError 对公开错误信息做 URL、域名和 API Key 脱敏。
func SanitizePublicError(err error) string {
	if err == nil {
		return "unknown error"
	}
	return SanitizePublicText(err.Error())
}

// SanitizePublicText 对公开文本做 URL、域名和 API Key 脱敏。
func SanitizePublicText(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "unknown error"
	}

	message = publicURLPattern.ReplaceAllString(message, "<redacted_url>")
	message = publicAPIKeyPattern.ReplaceAllStringFunc(message, func(value string) string {
		return MaskAPIKey(value)
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
