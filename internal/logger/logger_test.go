package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/config"
	"github.com/iyaki/ralph/internal/logger"
)

func TestNewLoggerDisabledByConfig(t *testing.T) {
	cfg := &config.Config{NoLog: true}
	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Enabled() {
		t.Fatal("expected logger to be disabled")
	}
	if l.File() != nil {
		t.Fatal("expected no file when disabled")
	}
}

func TestNewLoggerDoesNotApplyEnvOverridesDirectly(t *testing.T) {
	t.Setenv("RALPH_LOG_ENABLED", "0")
	cfg := &config.Config{NoLog: false}
	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !l.Enabled() {
		t.Fatal("expected logger to use resolved config and remain enabled")
	}
	if err := l.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestNewLoggerCreatesAndAppendsFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "ralph.log")
	cfg := &config.Config{NoLog: false, LogFile: logPath, LogTruncate: false}

	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !l.Enabled() || l.File() == nil {
		t.Fatal("expected logger to be enabled with file")
	}
	if err := l.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "Ralph run started") {
		t.Fatalf("expected log header in file, got %q", string(content))
	}
	if !strings.Contains(string(content), "Git branch:") {
		t.Fatalf("expected git branch line, got %q", string(content))
	}
}

func TestNewLoggerTruncatesWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "ralph.log")
	if err := os.WriteFile(logPath, []byte("old-content\n"), 0600); err != nil {
		t.Fatalf("failed to seed log file: %v", err)
	}

	cfg := &config.Config{NoLog: false, LogFile: logPath, LogTruncate: true}

	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if strings.Contains(string(content), "old-content") {
		t.Fatalf("expected old content to be truncated, got %q", string(content))
	}
}

func TestNewLoggerUsesTempFileWhenLogPathEmpty(t *testing.T) {
	cfg := &config.Config{NoLog: false, LogFile: ""}

	l, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.File() == nil {
		t.Fatal("expected temp log file to be created")
	}
	path := l.File().Name()
	if !strings.Contains(filepath.Base(path), "ralph-") {
		t.Fatalf("expected temp file naming pattern, got %q", path)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestCloseWithoutFile(t *testing.T) {
	l := &logger.Logger{}
	if err := l.Close(); err != nil {
		t.Fatalf("expected nil error for close without file, got %v", err)
	}
}
