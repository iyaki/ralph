package e2e_test

import (
	"testing"
)

func TestE2EInlinePrompt(t *testing.T) {
	tc := TestCase{
		Name: "Prompt Resolution: Inline Prompt",
		Args: []string{"--prompt", "This is an inline prompt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}

func TestE2EStdinPrompt(t *testing.T) {
	tc := TestCase{
		Name:  "Prompt Resolution: Stdin Prompt",
		Args:  []string{"-"}, // Hyphen indicates read from stdin explicitly
		Stdin: "This is a prompt from stdin",
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}

func TestE2EStdinPrompt_Implicit(t *testing.T) {
	tc := TestCase{
		Name:  "Prompt Resolution: Stdin Prompt (Implicit)",
		Args:  []string{}, // No args, should read from stdin implicitly if content is piped
		Stdin: "This is a prompt from stdin implicitly",
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}
