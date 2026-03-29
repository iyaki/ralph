// Package cli provides CLI commands and execution flow for Ralph.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/iyaki/ralph/internal/config"
)

const maxPositionalArgs = 2

// NewRalphCommand creates the root command for Ralph.
func NewRalphCommand() *cobra.Command {
	var cfg config.Config

	cmd := &cobra.Command{
		Use:   "ralph [options] [prompt] [scope]",
		Short: "POSIX-compliant AI Agentic Loop runner for spec-driven development",
		Long: `Ralph is a POSIX-compliant AI Agentic Loop shell runner aimed for spec-driven development workflows.
It loads prompts from files (with optional inline overrides) and comes with build/plan presets.

For extended documentation, examples, and configuration options, visit https://github.com/iyaki/ralph.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  ralph build
  ralph plan my-feature
  ralph --max-iterations 10 build
  ralph --prompt "Custom prompt text"
  echo "prompt from stdin" | ralph -`,
		Args: cobra.MaximumNArgs(maxPositionalArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommandLogic(cmd, args, &cfg)
		},
	}

	setupSharedFlags(cmd, &cfg)

	// Register subcommands
	cmd.AddCommand(NewInitCommand())
	cmd.AddCommand(NewRunCommand())

	return cmd
}
