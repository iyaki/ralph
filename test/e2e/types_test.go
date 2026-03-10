package e2e_test

// TestCase represents a single end-to-end test scenario for the Ralph CLI.
type TestCase struct {
	// Name is a human-readable identifier for the test scenario.
	Name string

	// Args are the command-line arguments passed to the Ralph binary.
	Args []string

	// Stdin is the input string to pipe to the process standard input.
	Stdin string

	// Env is a map of environment variables to set for the process.
	// These override the default environment.
	Env map[string]string

	// Files is a map of filename to content for fixture files to create
	// in the test's temporary working directory before execution.
	Files map[string]string

	// ExpectedExitCode is the expected integer exit code of the process.
	ExpectedExitCode int

	// ExpectedStdoutContains is a list of substrings that must appear in stdout.
	ExpectedStdoutContains []string

	// ExpectedStderrContains is a list of substrings that must appear in stderr.
	ExpectedStderrContains []string

	// ExpectedFiles is a list of filenames that are expected to exist
	// in the working directory after execution.
	ExpectedFiles []string

	// ExpectedFileContent is a map of filename to expected substrings that
	// must appear in the file content.
	ExpectedFileContent map[string][]string

	// ForbiddenOutput is a list of substrings that must NOT appear in stdout or stderr.
	ForbiddenOutput []string
}

// AgentFixture defines the configuration for the test-only agent used in E2E tests.
type AgentFixture struct {
	// Name is the identifier for the test agent (e.g., "ralph-test-agent").
	Name string

	// Behavior controls the mode of the test agent (e.g., "complete_once").
	// This corresponds to the RALPH_TEST_AGENT_MODE environment variable.
	Behavior string

	// ScriptPath is the absolute path to the compiled test agent binary.
	ScriptPath string
}

// RunResult captures the outcome of a single Ralph CLI execution.
type RunResult struct {
	// ExitCode is the process exit code.
	ExitCode int

	// Stdout is the captured standard output.
	Stdout string

	// Stderr is the captured standard error.
	Stderr string

	// DurationMs is the execution duration in milliseconds.
	DurationMs int64

	// CreatedFiles is a list of files present in the workdir after execution.
	CreatedFiles []string
}
