package notifier

// Builtins 是默认启用的零值可用通知器列表。
var Builtins = []Notifier{
	Console{},
	Telegram{},
	Webhook{},
}
