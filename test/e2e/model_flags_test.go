package e2e_test

import (
	"testing"
)

func TestE2EModelFlags(t *testing.T) {
	t.Run("ModelOverride", func(t *testing.T) {
		tc := TestCase{
			Name: "Override model via CLI flag",
			Args: []string{
				"--model", "test-model-123",
				"--max-iterations", "1",
				"--prompt", "hello",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"[ralph-test-agent] Args:",
				"--model",
				"test-model-123",
			},
		}
		runTestCase(t, tc)
	})

	t.Run("AgentModeOverride", func(t *testing.T) {
		tc := TestCase{
			Name: "Override agent mode via CLI flag",
			Args: []string{
				"--agent-mode", "test-mode-456",
				"--max-iterations", "1",
				"--prompt", "hello",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"[ralph-test-agent] Args:",
				"--agent",
				"test-mode-456",
			},
		}
		runTestCase(t, tc)
	})
}
