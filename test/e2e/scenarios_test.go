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

func TestE2EMaxIterations(t *testing.T) {
	tc := TestCase{
		Name: "Failure Path: Max Iterations Reached",
		Args: []string{"--prompt-file", "prompt.txt", "--max-iterations", "2"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "never_complete",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"max iterations reached",
		},
	}

	runTestCase(t, tc)
}

func TestE2EMissingPromptFile(t *testing.T) {
	tc := TestCase{
		Name: "Failure Path: Missing Prompt File",
		Args: []string{"--prompt-file", "non-existent-prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "never_complete",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"failed to read prompt file",
		},
	}

	runTestCase(t, tc)
}
