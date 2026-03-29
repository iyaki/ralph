package e2e_test

import "testing"

func TestE2EAgentSelection(t *testing.T) {
	t.Run("SelectClaudeAgent", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "select claude agent",
			Args: []string{
				"--agent", "claude",
				"--model", "claude-sonnet-4",
				"--max-iterations", "1",
				"--prompt", "hello",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"Using agent: claude",
				"[ralph-test-agent] Args:",
				"--dangerously-skip-permissions",
				"--model",
				"claude-sonnet-4",
			},
		})
	})

	t.Run("SelectCursorAgent", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "select cursor agent",
			Args: []string{
				"--agent", "cursor",
				"--max-iterations", "1",
				"--prompt", "hello",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"Using agent: cursor",
				"[ralph-test-agent] Args:",
			},
		})
	})

	t.Run("UnknownAgentReturnsError", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "unknown agent returns error",
			Args: []string{
				"--agent", "unknown-agent",
				"--max-iterations", "1",
				"--prompt", "hello",
			},
			ExpectedExitCode: 1,
			ExpectedStderrContains: []string{
				"unknown agent \"unknown-agent\"",
			},
		})
	})
}
