package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/config"
)

func writeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write executable: %v", err)
	}

	return path
}

func TestNewRalphCommandBasicProperties(t *testing.T) {
	cmd := NewRalphCommand()
	if !strings.Contains(cmd.Use, "ralph") {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Flags().Lookup("max-iterations") == nil {
		t.Fatal("expected max-iterations flag to exist")
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestNewRalphCommandExecuteDebugHappyPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", t.TempDir())
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	cmd := NewRalphCommand()
	cmd.SetArgs([]string{"build"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}
}

func TestNewRalphCommandExecuteConfigError(t *testing.T) {
	cmd := NewRalphCommand()
	cmd.SetArgs([]string{"--config", "missing-config.toml", "build"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected config loading error")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRalphCommandExecutePromptError(t *testing.T) {
	cmd := NewRalphCommand()
	cmd.SetArgs([]string{"--prompt-file", "missing-prompt.md", "build"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected prompt loading error")
	}
	if !strings.Contains(err.Error(), "failed to get prompt") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLoopCompletesOnSignal(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"$*\"\necho \"<promise>COMPLETE</promise>\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 3, AgentName: "opencode"}
	var out bytes.Buffer
	err := runLoop(cfg, "task <COMPLETION_SIGNAL>", "build", &out)
	if err != nil {
		t.Fatalf("expected completion success, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "All planned tasks completed!") {
		t.Fatalf("expected completion output, got %q", output)
	}
	if !strings.Contains(output, "<promise>COMPLETE</promise>") {
		t.Fatalf("expected replaced completion signal in agent input/output, got %q", output)
	}
}

func TestRunLoopDebugMode(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"should-not-run\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "1")

	cfg := &config.Config{MaxIterations: 2, AgentName: "opencode"}
	var out bytes.Buffer
	err := runLoop(cfg, "hello <COMPLETION_SIGNAL>", "plan", &out)
	if err != nil {
		t.Fatalf("expected debug mode to finish without error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "hello <promise>COMPLETE</promise>") {
		t.Fatalf("expected prompt with replaced signal in debug output, got %q", output)
	}
}

func TestRunLoopWarnsWhenAgentUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("DEBUG", "1")

	cfg := &config.Config{MaxIterations: 1, AgentName: "opencode"}
	var out bytes.Buffer
	err := runLoop(cfg, "debug", "build", &out)
	if err != nil {
		t.Fatalf("expected debug mode success, got %v", err)
	}

	if !strings.Contains(out.String(), "agent not found in PATH") {
		t.Fatalf("expected unavailable-agent warning, got %q", out.String())
	}
}

func TestRunLoopHandlesExecutionWarningAndMaxIterations(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"partial\"\nexit 1\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 1, AgentName: "opencode"}
	var out bytes.Buffer
	err := runLoop(cfg, "task", "build", &out)
	if err == nil {
		t.Fatal("expected max iterations error")
	}
	if !strings.Contains(err.Error(), "max iterations reached") {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Command execution warning") {
		t.Fatalf("expected execution warning, got %q", output)
	}
	if !strings.Contains(output, "Reached max iterations") {
		t.Fatalf("expected max iterations message, got %q", output)
	}
}

func TestRunLoopMaxIterationsWithoutCompletion(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"working\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 2, AgentName: "opencode"}
	err := runLoop(cfg, "task", "build", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected max iterations error")
	}
}
