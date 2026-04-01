package main

import (
	"dnsherene/internal/config"
	"fmt"
	"os"
)

// main 加载配置并执行一次续期任务。
func main() {
	summary := publicSummary{}

	cfg, err := config.Load()

	if err != nil {
		writePublicErrorReport(os.Stderr, err)
		os.Exit(1)
	}
	summary, err = run(cfg)

	if err != nil {
		writePublicErrorReport(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("renewed_total=%d\n", summary.Renewed)

}
