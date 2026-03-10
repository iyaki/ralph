package e2e_test

import (
	"testing"
)

func TestE2ELoggingFlags(t *testing.T) {
	t.Run("NoLog", func(t *testing.T) {
		tc := TestCase{
			Name: "Logging: Disabled via --no-log",
			Args: []string{"--log-file", "ralph.log", "--no-log", "--prompt-file", "prompt.txt"},
			Env: map[string]string{
				"RALPH_TEST_AGENT_MODE": "complete_once",
			},
			Files: map[string]string{
				"prompt.txt": "Just a simple prompt",
			},
			ExpectedExitCode: 0,
			ForbiddenFiles: []string{
				"ralph.log",
			},
		}

		runTestCase(t, tc)
	})

	t.Run("LogTruncate", func(t *testing.T) {
		tc := TestCase{
			Name: "Logging: Truncate",
			Args: []string{"--log-file", "ralph.log", "--log-truncate", "--prompt-file", "prompt.txt"},
			Env: map[string]string{
				"RALPH_TEST_AGENT_MODE": "complete_once",
			},
			Files: map[string]string{
				"prompt.txt": "Just a simple prompt",
				"ralph.log":  "OLD LOG CONTENT",
			},
			ExpectedExitCode: 0,
			ExpectedFiles: []string{
				"ralph.log",
			},
			ExpectedFileContent: map[string][]string{
				"ralph.log": {
					"===== Ralph run started at",
				},
			},
			ForbiddenFileContent: map[string][]string{
				"ralph.log": {
					"OLD LOG CONTENT",
				},
			},
		}

		runTestCase(t, tc)
	})
}
