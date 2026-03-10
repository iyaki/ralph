package e2e_test

import (
	"testing"
)

func TestE2ESpecsFlags(t *testing.T) {
	// 1. Test --specs-dir and --specs-index
	t.Run("Custom Specs Dir and Index", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "CustomSpecsDirAndIndex",
			Args: []string{
				"build",
				"--specs-dir", "custom-specs",
				"--specs-index", "INDEX.md",
			},
			Files: map[string]string{
				"custom-specs/INDEX.md": "# Custom Index File\nIndex content.",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"custom-specs/INDEX.md", // The agent should receive the index path in the prompt
			},
		})
	})

	// 2. Test --no-specs-index
	t.Run("No Specs Index", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "NoSpecsIndex",
			Args: []string{
				"build",
				"--no-specs-index",
			},
			Files: map[string]string{
				"specs/README.md": "# Default Index\nDefault index content.",
			},
			ExpectedExitCode: 0,
			ForbiddenOutput: []string{
				"specs/README.md", // Should NOT be present in the prompt
			},
		})
	})
}
