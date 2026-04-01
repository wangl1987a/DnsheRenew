package notifier

import (
	"context"
	"errors"
	"time"
)

type Level string

const (
	LevelInfo  Level = "info"
	LevelError Level = "error"
)

type Event struct {
	Level   Level          `json:"level"`
	Title   string         `json:"title"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields,omitempty"`
	Time    time.Time      `json:"time"`
}

// Notifier 抽象统一通知能力，业务层只依赖该接口。
type Notifier interface {
	Notify(ctx context.Context, event Event) error
}

// Multi 将同一事件扇出到多个通知实现（控制台、Webhook 等）。
type Multi struct {
	notifiers []Notifier
}

func NewMulti(notifiers ...Notifier) *Multi {
	filtered := make([]Notifier, 0, len(notifiers))
	for _, n := range notifiers {
		if n != nil {
			filtered = append(filtered, n)
		}
	}
	return &Multi{notifiers: filtered}
}

func (m *Multi) Notify(ctx context.Context, event Event) error {
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Notify(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
