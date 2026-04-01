package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Console struct {
	outWriter io.Writer
	errWriter io.Writer
}

func NewConsole(outWriter io.Writer, errWriter io.Writer) *Console {
	if outWriter == nil {
		outWriter = os.Stdout
	}
	if errWriter == nil {
		errWriter = os.Stderr
	}
	return &Console{
		outWriter: outWriter,
		errWriter: errWriter,
	}
}

func (c *Console) Notify(_ context.Context, event Event) error {
	// error 级别写 stderr，其余写 stdout，便于 CI/日志系统按流分类采集。
	writer := c.outWriter
	if event.Level == LevelError {
		writer = c.errWriter
	}

	title := strings.TrimSpace(event.Title)
	if title == "" {
		title = "Notification"
	}
	message := strings.TrimSpace(event.Message)
	if message == "" {
		message = "-"
	}
	ts := event.Time
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	header := fmt.Sprintf("[%s] [%s] %s: %s", ts.Format(time.RFC3339), strings.ToUpper(string(event.Level)), title, message)
	if _, err := fmt.Fprintln(writer, header); err != nil {
		return err
	}

	if len(event.Fields) == 0 {
		return nil
	}

	raw, err := json.Marshal(event.Fields)
	if err != nil {
		return fmt.Errorf("encode fields failed: %w", err)
	}
	if _, err := fmt.Fprintln(writer, string(raw)); err != nil {
		return err
	}
	return nil
}
