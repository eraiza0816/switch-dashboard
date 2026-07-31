package logbuf

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func discardingBuf(max int) *LogBuffer {
	return New(slog.NewTextHandler(io.Discard, nil), max, slog.LevelDebug)
}

func TestNewAndGet(t *testing.T) {
	b := discardingBuf(100)
	logger := slog.New(b)
	logger.Info("hello")
	logger.Warn("world", "key", "val")

	entries := b.Get()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "hello" {
		t.Fatalf("entry 0 message: got %q", entries[0].Message)
	}
	if entries[1].Message != "world" {
		t.Fatalf("entry 1 message: got %q", entries[1].Message)
	}
	if entries[1].Attrs["key"] != "val" {
		t.Fatalf("entry 1 attrs: got %v", entries[1].Attrs)
	}
}

func TestMaxEntries(t *testing.T) {
	b := discardingBuf(3)
	logger := slog.New(b)
	for i := 0; i < 10; i++ {
		logger.Info("msg")
	}
	entries := b.Get()
	if len(entries) > 3 {
		t.Fatalf("expected <= 3 entries, got %d", len(entries))
	}
}

func TestClear(t *testing.T) {
	b := discardingBuf(100)
	logger := slog.New(b)
	logger.Info("hello")
	b.Clear()
	if len(b.Get()) != 0 {
		t.Fatal("expected 0 entries after clear")
	}
}

func TestSetLevel(t *testing.T) {
	b := discardingBuf(100)
	logger := slog.New(b)

	logger.Debug("debug1")
	b.SetLevel(slog.LevelInfo)
	logger.Debug("debug2")
	logger.Info("info1")

	entries := b.Get()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (debug1 + info1), got %d", len(entries))
	}
	if entries[0].Message != "debug1" {
		t.Fatalf("entry 0: got %q", entries[0].Message)
	}
	if entries[1].Message != "info1" {
		t.Fatalf("entry 1: got %q", entries[1].Message)
	}
}

func TestLevel(t *testing.T) {
	b := discardingBuf(100)
	if b.Level() != slog.LevelDebug {
		t.Fatalf("default level: got %v", b.Level())
	}
	b.SetLevel(slog.LevelError)
	if b.Level() != slog.LevelError {
		t.Fatalf("after set: got %v", b.Level())
	}
}

func TestEnabled(t *testing.T) {
	b := discardingBuf(100)
	ctx := context.Background()
	if !b.Enabled(ctx, slog.LevelDebug) {
		t.Fatal("expected debug enabled")
	}
	b.SetLevel(slog.LevelWarn)
	if b.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("expected info disabled at warn level")
	}
	if !b.Enabled(ctx, slog.LevelError) {
		t.Fatal("expected error enabled at warn level")
	}
}

func TestGetReturnsCopy(t *testing.T) {
	b := discardingBuf(100)
	logger := slog.New(b)
	logger.Info("m1")

	entries := b.Get()
	entries[0].Message = "tampered"

	// Original should be unchanged
	orig := b.Get()
	if orig[0].Message != "m1" {
		t.Fatalf("original was mutated: got %q", orig[0].Message)
	}
}
