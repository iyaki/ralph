package agent

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write executable: %v", err)
	}
	return path
}

func TestGetAgentReturnsExpectedType(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		expected  string
	}{
		{name: "claude", agentName: "claude", expected: "claude"},
		{name: "cursor", agentName: "cursor", expected: "cursor"},
		{name: "opencode", agentName: "opencode", expected: "opencode"},
		{name: "default fallback", agentName: "unknown", expected: "opencode"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := GetAgent(tc.agentName, "model-x", "reviewer")
			if a.Name() != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, a.Name())
			}
		})
	}
}

func TestClaudeExecuteSuccessAndFailure(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "claude", "#!/bin/sh\necho \"out:$*\"\necho \"err:$*\" 1>&2\nif [ \"$FAIL\" = \"1\" ]; then exit 1; fi\n")
	t.Setenv("PATH", tmp)

	a := &ClaudeAgent{Model: "m1", AgentMode: "planner"}
	var out bytes.Buffer
	result, err := a.Execute("hello", &out)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !strings.Contains(result, "out:--dangerously-skip-permissions --model m1 --agent planner hello") {
		t.Fatalf("unexpected result: %q", result)
	}
	if !strings.Contains(out.String(), "err:") {
		t.Fatalf("expected stderr content in output writer: %q", out.String())
	}

	t.Setenv("FAIL", "1")
	_, err = a.Execute("hello", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "claude execution failed") {
		t.Fatalf("expected wrapped error, got %v", err)
	}

	t.Setenv("PATH", t.TempDir())
	if a.IsAvailable() {
		t.Fatal("expected claude to be unavailable")
	}
}

func TestCursorExecuteAndAvailability(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "cursor", "#!/bin/sh\necho \"cursor:$*\"\n")
	t.Setenv("PATH", tmp)

	a := &CursorAgent{Model: "m2"}
	if !a.IsAvailable() {
		t.Fatal("expected cursor to be available")
	}
	if a.Name() != "cursor" {
		t.Fatalf("unexpected name: %s", a.Name())
	}

	result, err := a.Execute("prompt", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "cursor:--model m2 prompt") {
		t.Fatalf("unexpected result: %q", result)
	}

	t.Setenv("PATH", t.TempDir())
	if a.IsAvailable() {
		t.Fatal("expected cursor to be unavailable")
	}
}

func TestOpencodeExecuteAndAvailability(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"open:$*\"\n")
	t.Setenv("PATH", tmp)

	a := &OpencodeAgent{Model: "m3", AgentMode: "reviewer"}
	if !a.IsAvailable() {
		t.Fatal("expected opencode to be available")
	}
	if a.Name() != "opencode" {
		t.Fatalf("unexpected name: %s", a.Name())
	}

	result, err := a.Execute("work", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "open:run --model m3 --agent reviewer work") {
		t.Fatalf("unexpected result: %q", result)
	}

	t.Setenv("PATH", t.TempDir())
	if a.IsAvailable() {
		t.Fatal("expected opencode to be unavailable")
	}
}
