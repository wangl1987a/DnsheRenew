package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dnsherene/internal/config"
	"dnsherene/internal/report"
)

const (
	telegramAPIBaseURL   = "https://api.telegram.org"
	telegramMessageLimit = 3500
)

type Telegram struct {
	config     config.TelegramConfig
	httpClient *http.Client
	apiBaseURL string
}

type telegramSendMessageRequest struct {
	ChatID          any    `json:"chat_id"`
	MessageThreadID *int   `json:"message_thread_id,omitempty"`
	Text            string `json:"text"`
	ParseMode       string `json:"parse_mode"`
}

type telegramSendMessageResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type telegramSection struct {
	Title      string
	Lines      []string
	Expandable bool
}

// Notify 在配置了 Telegram Bot 信息时发送美化后的结构化消息。
func (t Telegram) Notify(ctx context.Context, info report.Info) error {
	botToken := strings.TrimSpace(t.config.BotToken)
	chatID := t.config.ChatID

	if botToken == "" && chatID == 0 {
		return nil
	}
	if botToken == "" || chatID == 0 {
		return fmt.Errorf("telegram notifier requires bot token and chat id")
	}

	httpClient := t.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	apiBaseURL := strings.TrimRight(strings.TrimSpace(t.apiBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = telegramAPIBaseURL
	}

	endpoint := apiBaseURL + "/bot" + botToken + "/sendMessage"
	parseMode := strings.TrimSpace(t.config.ParseMode)
	if parseMode == "" {
		parseMode = "HTML"
	}

	for _, message := range buildTelegramMessages(info) {
		if err := t.sendMessage(
			ctx,
			httpClient,
			endpoint,
			chatID,
			t.config.MessageThreadID,
			message,
			parseMode,
		); err != nil {
			return err
		}
	}
	return nil
}

// sendMessage 调用 Telegram Bot API 的 sendMessage 方法。
func (t Telegram) sendMessage(
	ctx context.Context,
	httpClient *http.Client,
	endpoint string,
	chatID int64,
	messageThreadID *int,
	text string,
	parseMode string,
) error {
	payload := telegramSendMessageRequest{
		ChatID:          chatID,
		MessageThreadID: messageThreadID,
		Text:            text,
		ParseMode:       parseMode,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode telegram payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build telegram request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var out telegramSendMessageResponse
	_ = json.Unmarshal(body, &out)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram HTTP %d: %s", resp.StatusCode, firstNonEmpty(strings.TrimSpace(out.Description), strings.TrimSpace(string(body))))
	}
	if !out.OK {
		code := out.ErrorCode
		if code == 0 {
			code = resp.StatusCode
		}
		return fmt.Errorf("telegram API %d: %s", code, firstNonEmpty(strings.TrimSpace(out.Description), strings.TrimSpace(string(body))))
	}
	return nil
}

// buildTelegramMessages 将结构化报告渲染为适合 Telegram 的 HTML 消息列表。
func buildTelegramMessages(info report.Info) []string {
	blocks := []string{buildTelegramSummaryMessage(info)}
	for _, account := range info.Accounts {
		blocks = append(blocks, buildTelegramAccountMessages(account)...)
	}
	return packTelegramMessages(blocks)
}

// buildTelegramSummaryMessage 构建汇总消息。
func buildTelegramSummaryMessage(info report.Info) string {
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
		"生成时间: " + formatTelegramTime(ts),
		fmt.Sprintf("续期成功总数: <b>%d</b>", info.RenewedTotal),
		fmt.Sprintf("账号数量: <b>%d</b>", len(info.Accounts)),
		fmt.Sprintf("异常账号数: <b>%d</b>", failedAccounts),
	}

	return "<b>DNSHE 续期摘要</b>\n" + wrapTelegramBlockquote(lines, false)
}

// buildTelegramAccountMessages 按账号构建一条或多条 Telegram 消息。
func buildTelegramAccountMessages(account report.AccountInfo) []string {
	header := buildTelegramAccountHeader(account)
	sections := make([]telegramSection, 0, 3)

	if len(account.Domains) > 0 {
		sections = append(sections, telegramSection{
			Title:      "命中域名",
			Lines:      formatTelegramDomains(account.Domains),
			Expandable: len(account.Domains) > 8,
		})
	}
	if len(account.RenewedList) > 0 {
		sections = append(sections, telegramSection{
			Title:      "续期成功域名",
			Lines:      formatTelegramRenewed(account.RenewedList),
			Expandable: len(account.RenewedList) > 8,
		})
	}
	if len(account.FailedList) > 0 {
		sections = append(sections, telegramSection{
			Title:      "续期失败域名",
			Lines:      formatTelegramFailed(account.FailedList),
			Expandable: len(account.FailedList) > 5,
		})
	}

	if len(sections) == 0 {
		return []string{header}
	}

	messages := make([]string, 0, len(sections))
	current := header
	available := telegramMessageLimit - len(header) - 2
	if available < 500 {
		available = 500
	}

	for _, section := range sections {
		for _, piece := range splitTelegramSection(section, available) {
			rendered := renderTelegramSection(piece)
			if len(current)+2+len(rendered) <= telegramMessageLimit {
				current += "\n\n" + rendered
				continue
			}
			messages = append(messages, current)
			current = header + "\n\n" + rendered
		}
	}

	messages = append(messages, current)
	return messages
}

// packTelegramMessages 尽量把多个消息块合并为更少的 Telegram 消息。
func packTelegramMessages(blocks []string) []string {
	if len(blocks) == 0 {
		return nil
	}

	messages := make([]string, 0, len(blocks))
	current := ""
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		if current == "" {
			current = block
			continue
		}

		candidate := current + "\n\n" + block
		if len(candidate) <= telegramMessageLimit {
			current = candidate
			continue
		}

		messages = append(messages, current)
		current = block
	}

	if current != "" {
		messages = append(messages, current)
	}
	return messages
}

// buildTelegramAccountHeader 构建账号消息头部。
func buildTelegramAccountHeader(account report.AccountInfo) string {
	mask := fallback(account.APIKeyMasked, "***")
	lines := []string{
		fmt.Sprintf("命中续期窗口: <b>%d</b>", account.Matched),
		fmt.Sprintf("续期成功: <b>%d</b>", account.Renewed),
		fmt.Sprintf("续期失败: <b>%d</b>", account.Failed),
	}
	if account.DryRun {
		lines = append(lines, "演练模式: <b>是</b>")
	}
	if strings.TrimSpace(account.Error) != "" {
		lines = append(lines, "错误: "+html.EscapeString(account.Error))
	}

	return fmt.Sprintf(
		"<b>账号 [%d/%d]</b> <code>%s</code>\n%s",
		account.Index,
		account.Total,
		html.EscapeString(mask),
		wrapTelegramBlockquote(lines, false),
	)
}

// splitTelegramSection 按 Telegram 文本长度限制拆分区块。
func splitTelegramSection(section telegramSection, limit int) []telegramSection {
	if len(renderTelegramSection(section)) <= limit {
		return []telegramSection{section}
	}

	groups := make([][]string, 0, 1)
	current := make([]string, 0)
	for _, line := range section.Lines {
		candidate := append(append([]string{}, current...), line)
		if len(current) == 0 || len(renderTelegramSection(telegramSection{
			Title:      section.Title,
			Lines:      candidate,
			Expandable: section.Expandable,
		})) <= limit {
			current = candidate
			continue
		}
		groups = append(groups, current)
		current = []string{line}
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}

	if len(groups) <= 1 {
		return []telegramSection{section}
	}

	parts := make([]telegramSection, 0, len(groups))
	for i, group := range groups {
		parts = append(parts, telegramSection{
			Title:      fmt.Sprintf("%s (%d/%d)", section.Title, i+1, len(groups)),
			Lines:      group,
			Expandable: section.Expandable,
		})
	}
	return parts
}

// renderTelegramSection 渲染单个 Telegram 区块。
func renderTelegramSection(section telegramSection) string {
	return fmt.Sprintf(
		"<b>%s</b>\n%s",
		html.EscapeString(section.Title),
		wrapTelegramBlockquote(section.Lines, section.Expandable),
	)
}

// wrapTelegramBlockquote 使用 Telegram HTML blockquote 封装多行内容。
func wrapTelegramBlockquote(lines []string, expandable bool) string {
	if len(lines) == 0 {
		return "<blockquote>-</blockquote>"
	}

	tag := "blockquote"
	if expandable {
		tag = `blockquote expandable`
	}
	return "<" + tag + ">" + strings.Join(lines, "\n") + "</blockquote>"
}

// formatTelegramDomains 格式化域名列表。
func formatTelegramDomains(domains []report.DomainInfo) []string {
	lines := make([]string, 0, len(domains))
	for _, domain := range domains {
		line := "• <code>" + html.EscapeString(domain.Domain) + "</code>"
		meta := make([]string, 0, 2)
		if strings.TrimSpace(domain.ExpiresAt) != "" {
			meta = append(meta, "到期时间 <code>"+html.EscapeString(domain.ExpiresAt)+"</code>")
		}
		if domain.RemainingDays != nil {
			meta = append(meta, "剩余 <code>"+strconv.Itoa(*domain.RemainingDays)+" 天</code>")
		}
		if len(meta) > 0 {
			line += " — " + strings.Join(meta, " | ")
		}
		lines = append(lines, line)
	}
	return lines
}

// formatTelegramRenewed 格式化续期成功列表。
func formatTelegramRenewed(items []report.RenewedDomain) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := "• <code>" + html.EscapeString(item.Domain) + "</code>"
		if strings.TrimSpace(item.NewExpiresAt) != "" {
			line += " → <code>" + html.EscapeString(item.NewExpiresAt) + "</code>"
		}
		if item.RemainingDays > 0 {
			line += " (剩余 <code>" + strconv.Itoa(item.RemainingDays) + " 天</code>)"
		}
		lines = append(lines, line)
	}
	return lines
}

// formatTelegramFailed 格式化失败列表。
func formatTelegramFailed(items []report.FailedDomain) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		line := "• <code>" + html.EscapeString(item.Domain) + "</code>"
		if strings.TrimSpace(item.Reason) != "" {
			line += " — " + html.EscapeString(item.Reason)
		}
		lines = append(lines, line)
	}
	return lines
}

// formatTelegramTime 以 tg-time 标签输出时间。
func formatTelegramTime(ts time.Time) string {
	return fmt.Sprintf(
		`<tg-time unix="%d" format="wDT">%s</tg-time>`,
		ts.Unix(),
		html.EscapeString(ts.UTC().Format("2006-01-02 15:04:05 UTC")),
	)
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
