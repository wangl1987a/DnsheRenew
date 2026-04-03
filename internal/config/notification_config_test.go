package config

import "testing"

func TestLoadWithLookupReadsNotificationConfig(t *testing.T) {
	values := map[string]string{
		"DNSHE_API_KEYS":                  "k1,k2",
		"DNSHE_API_SECRETS":               "s1,s2",
		"DNSHE_DRY_RUN":                   "true",
		"DNSHE_DEBUG":                     "true",
		"DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN": "bot-123",
		"DNSHE_NOTIFY_TELEGRAM_CHAT_ID":   "-100123456",
	}

	cfg, err := loadWithLookup(func(key string) string {
		return values[key]
	})
	if err != nil {
		t.Fatalf("loadWithLookup returned error: %v", err)
	}
	if !cfg.Notification.Console.Enabled {
		t.Fatalf("expected Console.Enabled=true")
	}
	if cfg.Notification.Telegram == nil {
		t.Fatalf("expected Telegram config to be loaded")
	}
	if cfg.Notification.Telegram.ChatID != -100123456 {
		t.Fatalf("ChatID = %d, want -100123456", cfg.Notification.Telegram.ChatID)
	}
	if cfg.Notification.Webhook != nil {
		t.Fatalf("expected Webhook config to be nil when unset")
	}
}

func TestLoadWithLookupRejectsPartialTelegramConfig(t *testing.T) {
	values := map[string]string{
		"DNSHE_API_KEYS":                  "k1",
		"DNSHE_API_SECRETS":               "s1",
		"DNSHE_NOTIFY_TELEGRAM_BOT_TOKEN": "bot-123",
	}

	_, err := loadWithLookup(func(key string) string {
		return values[key]
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoadWithLookupReadsMailConfig(t *testing.T) {
	values := map[string]string{
		"DNSHE_API_KEYS":                  "k1",
		"DNSHE_API_SECRETS":               "s1",
		"DNSHE_NOTIFY_MAIL_FROM":          "sender@example.com",
		"DNSHE_NOTIFY_MAIL_SMTP_HOST":     "smtp.example.com",
		"DNSHE_NOTIFY_MAIL_SMTP_PORT":     "587",
		"DNSHE_NOTIFY_MAIL_SMTP_USERNAME": "user-1",
		"DNSHE_NOTIFY_MAIL_SMTP_PASSWORD": "pass-1",
		"DNSHE_NOTIFY_MAIL_TO":            "a@example.com,b@example.com",
	}

	cfg, err := loadWithLookup(func(key string) string {
		return values[key]
	})
	if err != nil {
		t.Fatalf("loadWithLookup returned error: %v", err)
	}
	if cfg.Notification.Mail == nil {
		t.Fatalf("expected Mail config to be loaded")
	}
	if cfg.Notification.Mail.SMTPAddress() != "smtp.example.com:587" {
		t.Fatalf("SMTPAddress = %q, want smtp.example.com:587", cfg.Notification.Mail.SMTPAddress())
	}
	if len(cfg.Notification.Mail.Receivers) != 2 {
		t.Fatalf("expected 2 mail receivers, got %d", len(cfg.Notification.Mail.Receivers))
	}
	if !cfg.Notification.Mail.HasAuth() {
		t.Fatalf("expected Mail.HasAuth=true")
	}
}

func TestLoadWithLookupReadsLarkWebhookConfig(t *testing.T) {
	values := map[string]string{
		"DNSHE_API_KEYS":                "k1",
		"DNSHE_API_SECRETS":             "s1",
		"DNSHE_NOTIFY_LARK_WEBHOOK_URL": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
	}

	cfg, err := loadWithLookup(func(key string) string {
		return values[key]
	})
	if err != nil {
		t.Fatalf("loadWithLookup returned error: %v", err)
	}
	if cfg.Notification.Lark == nil {
		t.Fatalf("expected Lark config to be loaded")
	}
	if cfg.Notification.Lark.Mode != LarkModeWebhook {
		t.Fatalf("Mode = %q, want %q", cfg.Notification.Lark.Mode, LarkModeWebhook)
	}
}

func TestLoadWithLookupRejectsPartialWebhookConfig(t *testing.T) {
	values := map[string]string{
		"DNSHE_API_KEYS":             "k1",
		"DNSHE_API_SECRETS":          "s1",
		"DNSHE_NOTIFY_WEBHOOK_TOKEN": "token-123",
	}

	_, err := loadWithLookup(func(key string) string {
		return values[key]
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
