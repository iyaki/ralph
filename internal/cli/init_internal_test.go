package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandWritesDefaultConfigFile(t *testing.T) {
	tmp := t.TempDir()

	oldTTYCheck := isInteractiveTerminal
	oldGetwd := getWorkingDir
	t.Cleanup(func() {
		isInteractiveTerminal = oldTTYCheck
		getWorkingDir = oldGetwd
	})

	isInteractiveTerminal = func() bool { return true }
	getWorkingDir = func() (string, error) { return tmp, nil }

	cmd := NewInitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected init to succeed, got %v", err)
	}

	configPath := filepath.Join(tmp, "ralph.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file at %s, got read error: %v", configPath, err)
	}

	contentText := string(content)
	if !strings.Contains(contentText, `agent = "opencode"`) {
		t.Fatalf("expected config to include default agent, got %q", contentText)
	}
}

func TestInitCommandWritesConfigToOutputPath(t *testing.T) {
	tmp := t.TempDir()
	targetPath := filepath.Join(tmp, "configs", "custom.toml")

	oldTTYCheck := isInteractiveTerminal
	t.Cleanup(func() {
		isInteractiveTerminal = oldTTYCheck
	})

	isInteractiveTerminal = func() bool { return true }

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"--output", targetPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected init to succeed, got %v", err)
	}

	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected config file at %s, got stat error: %v", targetPath, err)
	}
}
