package e2e_test

import (
	"testing"
)

func TestE2EPlanFlags(t *testing.T) {
	// Test --implementation-plan-name
	t.Run("Custom Implementation Plan Name", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "CustomImplementationPlanName",
			Args: []string{
				"build",
				"--implementation-plan-name", "CUSTOM_PLAN.md",
			},
			Env: map[string]string{
				"RALPH_TEST_AGENT_MODE": "complete_once",
			},
			Files: map[string]string{
				"CUSTOM_PLAN.md": "# Custom Plan\nSome content.",
			},
			ExpectedExitCode: 0,
			// Check stdout for the agent receiving the plan name in its arguments
			ExpectedStdoutContains: []string{
				"CUSTOM_PLAN.md",
			},
		})
	})
}
