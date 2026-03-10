package e2e_test

import (
	"testing"
)

func TestE2ECompletionFlow(t *testing.T) {
	tc := TestCase{
		Name: "Happy Path: Completion Detected",
		Args: []string{"--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}
