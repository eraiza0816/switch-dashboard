package logbuf

import (
	"context"
	"log/slog"
	"sync"
)

type Entry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

type LogBuffer struct {
	mu      sync.RWMutex
	entries []Entry
	max     int
	inner   slog.Handler
	level   slog.Leveler
}

func New(inner slog.Handler, max int, level slog.Leveler) *LogBuffer {
	return &LogBuffer{inner: inner, max: max, level: level}
}

func (b *LogBuffer) Get() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]Entry, len(b.entries))
	copy(result, b.entries)
	return result
}

func (b *LogBuffer) Clear() {
	b.mu.Lock()
	b.entries = nil
	b.mu.Unlock()
}

func (b *LogBuffer) SetLevel(lvl slog.Level) {
	b.mu.Lock()
	b.level = lvl
	b.mu.Unlock()
}

func (b *LogBuffer) Level() slog.Level {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.level.Level()
}

func (b *LogBuffer) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= b.Level()
}

func (b *LogBuffer) Handle(ctx context.Context, record slog.Record) error {
	if !b.Enabled(ctx, record.Level) {
		return nil
	}
	attrs := make(map[string]any)
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	entry := Entry{
		Time:    record.Time.Format("2006-01-02T15:04:05.000Z07:00"),
		Level:   record.Level.String(),
		Message: record.Message,
		Attrs:   attrs,
	}
	b.mu.Lock()
	b.entries = append(b.entries, entry)
	if len(b.entries) > b.max {
		b.entries = b.entries[len(b.entries)-b.max:]
	}
	b.mu.Unlock()
	return b.inner.Handle(ctx, record)
}

func (b *LogBuffer) WithAttrs(attrs []slog.Attr) slog.Handler {
	return b.inner.WithAttrs(attrs)
}

func (b *LogBuffer) WithGroup(name string) slog.Handler {
	return b.inner.WithGroup(name)
}
