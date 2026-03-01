package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// NewRalphCommand creates the root command for Ralph
func NewRalphCommand() *cobra.Command {
	var cfg Config

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
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse positional arguments
			promptName := "build"
			scope := "Whole system"

			if len(args) > 0 {
				promptName = args[0]
			}
			if len(args) > 1 {
				scope = args[1]
			}

			// Load configuration with proper precedence
			if err := cfg.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logger
			logger, err := NewLogger(&cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}
			defer logger.Close()

			// Write to both logger and stdout
			writers := []io.Writer{os.Stdout}
			if logger.enabled {
				writers = append(writers, logger.file)
			}
			output := io.MultiWriter(writers...)

			// Get the prompt
			prompt, err := GetPrompt(&cfg, promptName, scope, output)
			if err != nil {
				return fmt.Errorf("failed to get prompt: %w", err)
			}

			// Run the main loop
			return RunLoop(&cfg, prompt, promptName, output)
		},
	}

	// Setup flags
	flags := cmd.Flags()

	flags.StringVarP(&cfg.ConfigFile, "config", "c", "", "Config file to source")
	flags.IntVarP(&cfg.MaxIterations, "max-iterations", "m", 0, "Maximum iterations (default: 25)")
	flags.StringVarP(&cfg.PromptFile, "prompt-file", "p", "", "Prompt file path (use '-' to read from stdin)")
	flags.StringVarP(&cfg.SpecsDir, "specs-dir", "s", "", "Specs directory (default: specs)")
	flags.StringVarP(&cfg.SpecsIndexFile, "specs-index", "i", "", "Specs index file (default: README.md)")
	flags.BoolVar(&cfg.NoSpecsIndex, "no-specs-index", false, "Disable specs index file")
	flags.StringVarP(&cfg.ImplementationPlanName, "implementation-plan-name", "n", "", "Implementation plan file name")
	flags.StringVarP(&cfg.LogFile, "log-file", "l", "", "Log file path")
	flags.BoolVar(&cfg.NoLog, "no-log", false, "Disable logs")
	flags.BoolVar(&cfg.LogTruncate, "log-truncate", false, "Truncate log file before writing")
	flags.StringVar(&cfg.CustomPrompt, "prompt", "", "Inline custom prompt (overrides prompt files)")
	flags.StringVarP(&cfg.AgentName, "agent", "a", "", "AI agent to use: opencode, claude (default: opencode)")
	flags.StringVar(&cfg.Model, "model", "", "AI model to use (e.g., claude-sonnet-4, gpt-4)")
	flags.StringVar(&cfg.AgentMode, "agent-mode", "", "Agent mode/sub-agent to use (e.g., reviewer, planner)")

	return cmd
}

// RunLoop executes the main Ralph iteration loop
func RunLoop(cfg *Config, prompt, promptName string, output io.Writer) error {
	completionSignal := "<promise>COMPLETE</promise>"

	// Replace placeholders in prompt
	prompt = strings.ReplaceAll(prompt, "<COMPLETION_SIGNAL>", completionSignal)

	// Get the configured agent
	agent := GetAgent(cfg.AgentName, cfg.Model, cfg.AgentMode)

	// Check if agent is available
	if !agent.IsAvailable() {
		fmt.Fprintf(output, "Warning: %s agent not found in PATH, will continue anyway...\n", agent.Name())
	}

	fmt.Fprintf(output, "Starting Ralph - Max iterations: %d\n", cfg.MaxIterations)
	fmt.Fprintf(output, "Using agent: %s\n", agent.Name())

	for i := 1; i <= cfg.MaxIterations; i++ {
		fmt.Fprintf(output, "\n")
		fmt.Fprintf(output, "===============================================================\n")
		fmt.Fprintf(output, " [%s] Iteration %d of %d (%s)\n", promptName, i, cfg.MaxIterations, time.Now().Format(time.RFC3339))
		fmt.Fprintf(output, "===============================================================\n")

		// Check if DEBUG mode (for testing)
		if os.Getenv("DEBUG") != "" {
			fmt.Fprintln(output, prompt)
			fmt.Fprintf(output, "\nAll planned tasks completed!\n")
			fmt.Fprintf(output, "Completed at iteration %d of %d\n", i, cfg.MaxIterations)
			return nil
		}

		// Execute the agent
		result, err := agent.Execute(prompt, output)
		if err != nil {
			// Non-fatal error, continue to next iteration
			fmt.Fprintf(output, "Command execution warning: %v\n", err)
		}

		// Check for completion signal
		if strings.Contains(result, completionSignal) {
			fmt.Fprintf(output, "\nAll planned tasks completed!\n")
			fmt.Fprintf(output, "Completed at iteration %d of %d\n", i, cfg.MaxIterations)
			return nil
		}

		fmt.Fprintf(output, "Iteration %d complete. Continuing...\n", i)
	}

	fmt.Fprintf(output, "\nReached max iterations (%d) without completing all planned tasks.\n", cfg.MaxIterations)
	return fmt.Errorf("max iterations reached")
}
