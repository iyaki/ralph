package e2e_test

import "testing"

const testAgentEnvEchoKeys = "RALPH_TEST_AGENT_ECHO_ENV_KEYS"

func TestE2EEnvOverrides(t *testing.T) {
	testCases := []struct {
		name string
		tc   TestCase
	}{
		{name: "FlagOnlyOverride", tc: envFlagOnlyOverrideCase()},
		{name: "ConfigOnlyOverride", tc: envConfigOnlyOverrideCase()},
		{name: "FlagOverridesConfig", tc: envFlagOverridesConfigCase()},
		{name: "RepeatedFlagLastWins", tc: envRepeatedFlagLastWinsCase()},
		{name: "InvalidEntryFailsBeforeExecution", tc: envInvalidEntryCase()},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runTestCase(t, testCase.tc)
		})
	}
}

func envFlagOnlyOverrideCase() TestCase {
	return TestCase{
		Name: "Env overrides: flag-only",
		Args: []string{
			"--max-iterations", "1",
			"--prompt", "hello",
			"--env", "RALPH_E2E_CHILD_ONLY=from-flag",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			testAgentEnvEchoKeys:    "RALPH_E2E_CHILD_ONLY",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Env RALPH_E2E_CHILD_ONLY=from-flag",
		},
	}
}

func envConfigOnlyOverrideCase() TestCase {
	return TestCase{
		Name: "Env overrides: config-only",
		Args: []string{
			"--config", "ralph.toml",
			"--max-iterations", "1",
			"--prompt", "hello",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			testAgentEnvEchoKeys:    "RALPH_E2E_CHILD_ONLY",
		},
		Files: map[string]string{
			"ralph.toml": "[env]\nRALPH_E2E_CHILD_ONLY = \"from-config\"\n",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Env RALPH_E2E_CHILD_ONLY=from-config",
		},
	}
}

func envFlagOverridesConfigCase() TestCase {
	return TestCase{
		Name: "Env overrides: flag overrides config",
		Args: []string{
			"--config", "ralph.toml",
			"--max-iterations", "1",
			"--prompt", "hello",
			"--env", "RALPH_E2E_PRECEDENCE=from-flag",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			testAgentEnvEchoKeys:    "RALPH_E2E_PRECEDENCE",
		},
		Files: map[string]string{
			"ralph.toml": "[env]\nRALPH_E2E_PRECEDENCE = \"from-config\"\n",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Env RALPH_E2E_PRECEDENCE=from-flag",
		},
		ForbiddenOutput: []string{
			"[ralph-test-agent] Env RALPH_E2E_PRECEDENCE=from-config",
		},
	}
}

func envRepeatedFlagLastWinsCase() TestCase {
	return TestCase{
		Name: "Env overrides: repeated flag last wins",
		Args: []string{
			"--max-iterations", "1",
			"--prompt", "hello",
			"--env", "RALPH_E2E_DUPLICATE=one",
			"--env", "RALPH_E2E_DUPLICATE=two",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			testAgentEnvEchoKeys:    "RALPH_E2E_DUPLICATE",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Env RALPH_E2E_DUPLICATE=two",
		},
		ForbiddenOutput: []string{
			"[ralph-test-agent] Env RALPH_E2E_DUPLICATE=one",
		},
	}
}

func envInvalidEntryCase() TestCase {
	return TestCase{
		Name: "Env overrides: invalid entry fails before execution",
		Args: []string{
			"--max-iterations", "1",
			"--prompt", "hello",
			"--env", "1INVALID=super-secret-token",
		},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"invalid --env key",
		},
		ForbiddenOutput: []string{
			"super-secret-token",
			"[ralph-test-agent] Starting",
		},
	}
}
