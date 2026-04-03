package main

import (
	"context"
	"dnsherene/internal/config"
	"dnsherene/internal/notification"
	"dnsherene/internal/output"
	"dnsherene/internal/runner"
	"fmt"
	"os"
)

// main 加载配置并执行一次续期任务。
func main() {
	cfg, err := config.Load()
	if err != nil {
		output.WritePublicErrorReport(os.Stderr, err)
		os.Exit(1)
	}

	notifier, err := notification.NewManager(cfg.Notification)
	if err != nil {
		output.WritePublicErrorReport(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	info, err := runner.Execute(ctx, cfg)
	_ = notifier.Notify(ctx, info)
	if err != nil {
		output.WritePublicErrorReport(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("renewed_total=%d\n", info.RenewedTotal)
}
