package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type runTestCase struct {
	name           string
	mode           string
	expectedExit   int
	expectedOutput string
	expectedErr    string
}

func TestRun(t *testing.T) {
	tests := []runTestCase{
		{
			name:           "complete_once",
			mode:           modeCompleteOnce,
			expectedExit:   exitCodeSuccess,
			expectedOutput: "<promise>COMPLETE</promise>",
		},
		{
			name:           "never_complete",
			mode:           modeNeverComplete,
			expectedExit:   exitCodeSuccess,
			expectedOutput: "Processing request forever...",
		},
		{
			name:         "return_error",
			mode:         modeReturnError,
			expectedExit: exitCodeError,
			expectedErr:  "Simulated agent failure",
		},
		{
			name:           "slow_complete",
			mode:           modeSlowComplete,
			expectedExit:   exitCodeSuccess,
			expectedOutput: "<promise>COMPLETE</promise>",
		},
		{
			name:         "unknown_mode",
			mode:         "unknown",
			expectedExit: exitCodeUnknown,
			expectedErr:  "Unknown mode: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt)
		})
	}
}

func runTest(t *testing.T, tt runTestCase) {
	t.Helper()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	getEnv := func(key string) string {
		if key == "RALPH_TEST_AGENT_MODE" {
			return tt.mode
		}

		return ""
	}

	start := time.Now()
	exitCode := run([]string{"prog", "arg"}, getEnv, stdout, stderr)
	duration := time.Since(start)

	if exitCode != tt.expectedExit {
		t.Errorf("expected exit code %d, got %d", tt.expectedExit, exitCode)
	}

	if tt.expectedOutput != "" && !strings.Contains(stdout.String(), tt.expectedOutput) {
		t.Errorf("expected stdout to contain %q, got %q", tt.expectedOutput, stdout.String())
	}

	if tt.expectedErr != "" && !strings.Contains(stderr.String(), tt.expectedErr) {
		t.Errorf("expected stderr to contain %q, got %q", tt.expectedErr, stderr.String())
	}

	if tt.mode == modeSlowComplete && duration < 50*time.Millisecond {
		t.Errorf("expected slow_complete to take at least 50ms, took %v", duration)
	}
}
