package config

import (
	"fmt"
	"strconv"
	"strings"
)

// NotificationConfig 表示全部通知渠道配置。
type NotificationConfig struct {
	Console  ConsoleConfig
	Telegram *TelegramConfig
	Mail     *MailConfig
	Lark     *LarkConfig
	Webhook  *WebhookConfig
}

// ConsoleConfig 表示控制台通知配置。
type ConsoleConfig struct {
	Enabled bool
}

// TelegramConfig 表示 Telegram 通知配置。
type TelegramConfig struct {
	BotToken        string
	ChatID          int64
	MessageThreadID *int
	ParseMode       string
}

// MailConfig 表示邮件通知配置。
type MailConfig struct {
	SenderAddress string
	SMTPHost      string
	SMTPPort      int
	SMTPIdentity  string
	SMTPUsername  string
	SMTPPassword  string
	Receivers     []string
}

// SMTPAddress 返回 SMTP 服务地址。
func (c MailConfig) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}

// HasAuth 返回邮件通知是否配置了 SMTP 认证。
func (c MailConfig) HasAuth() bool {
	return strings.TrimSpace(c.SMTPIdentity) != "" ||
		strings.TrimSpace(c.SMTPUsername) != "" ||
		strings.TrimSpace(c.SMTPPassword) != ""
}

// LarkMode 表示 Lark 通知的发送模式。
type LarkMode string

const (
	// LarkModeWebhook 使用群机器人 webhook。
	LarkModeWebhook LarkMode = "webhook"
	// LarkModeCustomApp 使用自建应用 + 接收人 ID。
	LarkModeCustomApp LarkMode = "custom_app"
)

// LarkConfig 表示 Lark 通知配置。
type LarkConfig struct {
	Mode         LarkMode
	WebhookURL   string
	AppID        string
	AppSecret    string
	ReceiverType string
	Receivers    []string
}

// WebhookConfig 表示自定义 webhook 通知配置。
type WebhookConfig struct {
	URL   string
	Token string
}

func loadNotificationConfig(lookup func(string) string) (NotificationConfig, error) {
	cfg := NotificationConfig{
		Console: ConsoleConfig{
			Enabled: parseBool(lookup("DNSHE_DEBUG")),
		},
	}

	var err error
	cfg.Telegram, err = loadTelegramConfig(lookup)
	if err != nil {
		return cfg, err
	}

	cfg.Mail, err = loadMailConfig(lookup)
	if err != nil {
		return cfg, err
	}

	cfg.Lark, err = loadLarkConfig(lookup)
	if err != nil {
		return cfg, err
	}

	cfg.Webhook, err = loadWebhookConfig(lookup)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func loadTelegramConfig(lookup func(string) string) (*TelegramConfig, error) {
	botToken := strings.TrimSpace(lookup("DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN"))
	chatIDRaw := strings.TrimSpace(lookup("DNSHE_NOTIFY_TELEGRAM_CHAT_ID"))
	messageThreadIDRaw := strings.TrimSpace(lookup("DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID"))
	parseMode := strings.TrimSpace(lookup("DNSHE_NOTIFY_TELEGRAM_PARSE_MODE"))

	// Telegram 是否启用只由 bot token 和 chat id 决定。
	// parse mode / message thread id 都是附加选项，不应单独触发配置错误。
	if allEmpty(botToken, chatIDRaw) {
		return nil, nil
	}
	if botToken == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN is required when Telegram notification is configured")
	}
	if chatIDRaw == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_TELEGRAM_CHAT_ID is required when Telegram notification is configured")
	}

	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("DNSHE_NOTIFY_TELEGRAM_CHAT_ID must be an integer")
	}

	messageThreadID, err := parseOptionalPositiveInt(
		messageThreadIDRaw,
		"DNSHE_NOTIFY_TELEGRAM_MESSAGE_THREAD_ID",
	)
	if err != nil {
		return nil, err
	}

	return &TelegramConfig{
		BotToken:        botToken,
		ChatID:          chatID,
		MessageThreadID: messageThreadID,
		ParseMode:       parseMode,
	}, nil
}

func loadMailConfig(lookup func(string) string) (*MailConfig, error) {
	senderAddress := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_FROM"))
	smtpHost := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_SMTP_HOST"))
	smtpPortRaw := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_SMTP_PORT"))
	smtpIdentity := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_SMTP_IDENTITY"))
	smtpUsername := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_SMTP_USERNAME"))
	smtpPassword := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_SMTP_PASSWORD"))
	receiversRaw := strings.TrimSpace(lookup("DNSHE_NOTIFY_MAIL_TO"))
	receivers := splitList(receiversRaw)

	if allEmpty(
		senderAddress,
		smtpHost,
		smtpPortRaw,
		smtpIdentity,
		smtpUsername,
		smtpPassword,
		receiversRaw,
	) {
		return nil, nil
	}
	if senderAddress == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_FROM is required when mail notification is configured")
	}
	if smtpHost == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_SMTP_HOST is required when mail notification is configured")
	}
	if smtpPortRaw == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_SMTP_PORT is required when mail notification is configured")
	}
	if len(receivers) == 0 {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_TO is required when mail notification is configured")
	}

	smtpPort, err := strconv.Atoi(smtpPortRaw)
	if err != nil || smtpPort <= 0 {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_SMTP_PORT must be a positive integer")
	}

	hasAnyAuthField := !allEmpty(smtpIdentity, smtpUsername, smtpPassword)
	if hasAnyAuthField && (smtpUsername == "" || smtpPassword == "") {
		return nil, fmt.Errorf("DNSHE_NOTIFY_MAIL_SMTP_USERNAME and DNSHE_NOTIFY_MAIL_SMTP_PASSWORD must be set together")
	}

	return &MailConfig{
		SenderAddress: senderAddress,
		SMTPHost:      smtpHost,
		SMTPPort:      smtpPort,
		SMTPIdentity:  smtpIdentity,
		SMTPUsername:  smtpUsername,
		SMTPPassword:  smtpPassword,
		Receivers:     receivers,
	}, nil
}

func loadLarkConfig(lookup func(string) string) (*LarkConfig, error) {
	webhookURL := strings.TrimSpace(lookup("DNSHE_NOTIFY_LARK_WEBHOOK_URL"))
	appID := strings.TrimSpace(lookup("DNSHE_NOTIFY_LARK_APP_ID"))
	appSecret := strings.TrimSpace(lookup("DNSHE_NOTIFY_LARK_APP_SECRET"))
	receiverType := strings.TrimSpace(lookup("DNSHE_NOTIFY_LARK_RECEIVER_TYPE"))
	receiversRaw := strings.TrimSpace(lookup("DNSHE_NOTIFY_LARK_RECEIVERS"))
	receivers := splitList(receiversRaw)

	if allEmpty(webhookURL, appID, appSecret, receiverType, receiversRaw) {
		return nil, nil
	}

	if webhookURL != "" {
		if !allEmpty(appID, appSecret, receiverType, receiversRaw) {
			return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_WEBHOOK_URL cannot be combined with custom app settings")
		}
		return &LarkConfig{
			Mode:       LarkModeWebhook,
			WebhookURL: webhookURL,
		}, nil
	}

	if appID == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_APP_ID is required when Lark custom app notification is configured")
	}
	if appSecret == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_APP_SECRET is required when Lark custom app notification is configured")
	}
	if receiverType == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_RECEIVER_TYPE is required when Lark custom app notification is configured")
	}
	if len(receivers) == 0 {
		return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_RECEIVERS is required when Lark custom app notification is configured")
	}
	if !isSupportedLarkReceiverType(receiverType) {
		return nil, fmt.Errorf("DNSHE_NOTIFY_LARK_RECEIVER_TYPE must be one of: open_id, user_id, union_id, email, chat_id")
	}

	return &LarkConfig{
		Mode:         LarkModeCustomApp,
		AppID:        appID,
		AppSecret:    appSecret,
		ReceiverType: receiverType,
		Receivers:    receivers,
	}, nil
}

func loadWebhookConfig(lookup func(string) string) (*WebhookConfig, error) {
	url := strings.TrimSpace(lookup("DNSHE_NOTIFY_WEBHOOK_URL"))
	token := strings.TrimSpace(lookup("DNSHE_NOTIFY_WEBHOOK_TOKEN"))

	if allEmpty(url, token) {
		return nil, nil
	}
	if url == "" {
		return nil, fmt.Errorf("DNSHE_NOTIFY_WEBHOOK_URL is required when DNSHE_NOTIFY_WEBHOOK_TOKEN is set")
	}

	return &WebhookConfig{
		URL:   url,
		Token: token,
	}, nil
}

func parseOptionalPositiveInt(raw string, key string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be an integer", key)
	}
	if value <= 0 {
		return nil, fmt.Errorf("%s must be positive", key)
	}
	return &value, nil
}

func allEmpty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func isSupportedLarkReceiverType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open_id", "user_id", "union_id", "email", "chat_id":
		return true
	default:
		return false
	}
}
