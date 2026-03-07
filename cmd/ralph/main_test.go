package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRunHelpReturnsZero(t *testing.T) {
	var errBuf bytes.Buffer
	exitCode := run([]string{"--help"}, &errBuf)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
}

func TestRunInvalidFlagReturnsOne(t *testing.T) {
	var errBuf bytes.Buffer
	exitCode := run([]string{"--unknown-flag"}, &errBuf)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(errBuf.String(), "Error:") {
		t.Fatalf("expected error output, got %q", errBuf.String())
	}
}

func TestMainProcessHelpExitCode(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainProcessHelper", "--", "--help")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected main helper process to exit 0, got: %v", err)
	}
}

func TestMainProcessHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := []string{}
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	os.Args = append([]string{"ralph-test"}, args...)
	main()
}
