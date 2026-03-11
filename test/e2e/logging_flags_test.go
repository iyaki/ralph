package e2e_test

import "testing"

func TestE2ELoggingFlags(t *testing.T) {
	testCases := []struct {
		name string
		tc   TestCase
	}{
		{name: "DefaultNoLog", tc: loggingDefaultNoLogCase()},
		{name: "EnabledViaEnv", tc: loggingEnabledViaEnvCase()},
		{name: "NoLogFalseOverridesConfig", tc: loggingNoLogFalseOverridesConfigCase()},
		{name: "NoLogFalseOverridesEnv", tc: loggingNoLogFalseOverridesEnvCase()},
		{name: "NoLog", tc: loggingNoLogFlagCase()},
		{name: "LogTruncate", tc: loggingTruncateCase()},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(t, testCase.tc)
		})
	}
}

func loggingDefaultNoLogCase() TestCase {
	return TestCase{
		Name: "Logging: Disabled by default",
		Args: []string{"--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ForbiddenFiles:   []string{"ralph.log"},
	}
}

func loggingEnabledViaEnvCase() TestCase {
	return TestCase{
		Name: "Logging: Enabled via RALPH_LOG_ENABLED=1",
		Args: []string{"--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_LOG_ENABLED":     "1",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ExpectedFiles:    []string{"ralph.log"},
	}
}

func loggingNoLogFalseOverridesConfigCase() TestCase {
	return TestCase{
		Name: "Logging: --no-log=false overrides config",
		Args: []string{
			"--config", "ralph.toml",
			"--prompt-file", "prompt.txt",
			"--log-file", "ralph.log",
			"--no-log=false",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
			"ralph.toml": "no-log = true\n",
		},
		ExpectedExitCode: 0,
		ExpectedFiles:    []string{"ralph.log"},
	}
}

func loggingNoLogFalseOverridesEnvCase() TestCase {
	return TestCase{
		Name: "Logging: --no-log=false overrides env",
		Args: []string{
			"--prompt-file", "prompt.txt",
			"--log-file", "ralph.log",
			"--no-log=false",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_LOG_ENABLED":     "0",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ExpectedFiles:    []string{"ralph.log"},
	}
}

func loggingNoLogFlagCase() TestCase {
	return TestCase{
		Name: "Logging: Disabled via --no-log",
		Args: []string{"--log-file", "ralph.log", "--no-log", "--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ForbiddenFiles:   []string{"ralph.log"},
	}
}

func loggingTruncateCase() TestCase {
	return TestCase{
		Name: "Logging: Truncate",
		Args: []string{"--log-file", "ralph.log", "--log-truncate", "--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_LOG_ENABLED":     "1",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
			"ralph.log":  "OLD LOG CONTENT",
		},
		ExpectedExitCode: 0,
		ExpectedFiles:    []string{"ralph.log"},
		ExpectedFileContent: map[string][]string{
			"ralph.log": {"===== Ralph run started at"},
		},
		ForbiddenFileContent: map[string][]string{
			"ralph.log": {"OLD LOG CONTENT"},
		},
	}
}
