package executor_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iyaki/ralph/internal/executor"
)

func TestExecuteCommandSuccess(t *testing.T) {
	var out bytes.Buffer
	result, err := executor.ExecuteCommand("sh", []string{"-c", "echo out && echo err 1>&2"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "out") || !strings.Contains(result, "err") {
		t.Fatalf("unexpected result: %q", result)
	}
	if !strings.Contains(out.String(), "out") || !strings.Contains(out.String(), "err") {
		t.Fatalf("expected writer output to contain stdout and stderr; got %q", out.String())
	}
}

func TestExecuteCommandFailureReturnsOutputAndError(t *testing.T) {
	result, err := executor.ExecuteCommand("sh", []string{"-c", "echo partial && exit 2"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "command execution failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "partial") {
		t.Fatalf("expected partial output to be returned, got %q", result)
	}
}
