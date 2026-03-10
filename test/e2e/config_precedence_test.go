package e2e_test

import (
	"testing"
)

// TestE2EConfigPrecedence_FlagWins verifies that CLI flags take precedence over
// environment variables and config files.
func TestE2EConfigPrecedence_FlagWins(t *testing.T) {
	tc := TestCase{
		Name: "ConfigPrecedence_FlagWins",
		Args: []string{"--max-iterations", "15"},
		Env: map[string]string{
			"RALPH_MAX_ITERATIONS": "10",
		},
		Files: map[string]string{
			"ralph.toml": `max-iterations = 5`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 15",
		},
	}
	runTestCase(t, tc)
}

// TestE2EConfigPrecedence_EnvWins verifies that environment variables take precedence
// over config files when no flag is provided.
func TestE2EConfigPrecedence_EnvWins(t *testing.T) {
	tc := TestCase{
		Name: "ConfigPrecedence_EnvWins",
		Args: []string{},
		Env: map[string]string{
			"RALPH_MAX_ITERATIONS": "10",
		},
		Files: map[string]string{
			"ralph.toml": `max-iterations = 5`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 10",
		},
	}
	runTestCase(t, tc)
}

// TestE2EConfigPrecedence_ConfigFileWins verifies that config file values are used
// when no flag or environment variable is provided.
func TestE2EConfigPrecedence_ConfigFileWins(t *testing.T) {
	tc := TestCase{
		Name: "ConfigPrecedence_ConfigFileWins",
		Args: []string{},
		Env:  map[string]string{}, // No Env var
		Files: map[string]string{
			"ralph.toml": `max-iterations = 5`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 5",
		},
	}
	runTestCase(t, tc)
}

// TestE2EConfigPrecedence_RalphConfigEnvOverride verifies that RALPH_CONFIG environment
// variable can override the default config file path.
func TestE2EConfigPrecedence_RalphConfigEnvOverride(t *testing.T) {
	tc := TestCase{
		Name: "ConfigPrecedence_RalphConfigEnvOverride",
		Args: []string{},
		Env: map[string]string{
			"RALPH_CONFIG": "custom_config.toml",
		},
		Files: map[string]string{
			"custom_config.toml": `max-iterations = 7`,
			"ralph.toml":         `max-iterations = 5`, // Should be ignored
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 7",
		},
	}
	runTestCase(t, tc)
}

// TestE2EConfigPrecedence_ConfigFlagOverride verifies that --config flag can override
// both default config file path and RALPH_CONFIG environment variable.
func TestE2EConfigPrecedence_ConfigFlagOverride(t *testing.T) {
	tc := TestCase{
		Name: "ConfigPrecedence_ConfigFlagOverride",
		Args: []string{"--config", "flag_config.toml"},
		Env: map[string]string{
			"RALPH_CONFIG": "env_config.toml", // Should be ignored
		},
		Files: map[string]string{
			"flag_config.toml": `max-iterations = 8`,
			"env_config.toml":  `max-iterations = 7`,
			"ralph.toml":       `max-iterations = 5`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 8",
		},
	}
	runTestCase(t, tc)
}
