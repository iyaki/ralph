// Package cli provides CLI commands and execution flow for Ralph.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iyaki/ralph/internal/agent"
	"github.com/iyaki/ralph/internal/config"
	"github.com/iyaki/ralph/internal/logger"
	"github.com/iyaki/ralph/internal/prompt"
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
			promptName, scope := parsePositionalArgs(args)

			// Load configuration with proper precedence
			if err := cfg.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logger
			appLogger, err := logger.NewLogger(&cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}
			defer func() {
				_ = appLogger.Close()
			}()

			// Write to both logger and stdout
			writers := []io.Writer{cmd.OutOrStdout()}
			if appLogger.Enabled() {
				writers = append(writers, appLogger.File())
			}
			output := io.MultiWriter(writers...)

			// Get the prompt
			promptText, fmOverride, err := prompt.GetPrompt(&cfg, promptName, scope, output)
			if err != nil {
				return fmt.Errorf("failed to get prompt: %w", err)
			}

			// Apply configuration precedence
			applyEffectiveSettings(&cfg, cmd, fmOverride, promptName)

			// Run the main loop
			return RunLoop(&cfg, promptText, promptName, output)
		},
	}

	setupRootFlags(cmd, &cfg)

	// Register subcommands
	cmd.AddCommand(NewInitCommand())

	return cmd
}

func applyEffectiveSettings(
	cfg *config.Config,
	cmd *cobra.Command,
	fmOverride *config.PromptConfigOverride,
	promptName string,
) {
	applyModelSettings(cfg, cmd, fmOverride, promptName)
	applyAgentModeSettings(cfg, cmd, fmOverride, promptName)
}

func applyModelSettings(
	cfg *config.Config,
	cmd *cobra.Command,
	fmOverride *config.PromptConfigOverride,
	promptName string,
) {
	// Model Precedence: Flag > Env > Front Matter > Config Override > Global Config
	if cmd.Flags().Changed("model") || os.Getenv("RALPH_MODEL") != "" {
		return
	}

	if fmOverride != nil && fmOverride.Model != "" {
		cfg.Model = fmOverride.Model

		return
	}

	if promptOverride, ok := cfg.PromptOverrides[promptName]; ok && promptOverride.Model != "" {
		cfg.Model = promptOverride.Model
	}
}

func applyAgentModeSettings(
	cfg *config.Config,
	cmd *cobra.Command,
	fmOverride *config.PromptConfigOverride,
	promptName string,
) {
	// Agent Mode Precedence: Flag > Env > Front Matter > Config Override > Global Config
	if cmd.Flags().Changed("agent-mode") || os.Getenv("RALPH_AGENT_MODE") != "" {
		return
	}

	if fmOverride != nil && fmOverride.AgentMode != "" {
		cfg.AgentMode = fmOverride.AgentMode

		return
	}

	if promptOverride, ok := cfg.PromptOverrides[promptName]; ok && promptOverride.AgentMode != "" {
		cfg.AgentMode = promptOverride.AgentMode
	}
}

func parsePositionalArgs(args []string) (string, string) {
	promptName := "build"
	scope := "Whole system"

	if len(args) > 0 {
		promptName = args[0]
	}
	if len(args) > 1 {
		scope = args[1]
	}

	return promptName, scope
}

func setupRootFlags(cmd *cobra.Command, cfg *config.Config) {
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
	flags.StringVarP(&cfg.AgentName, "agent", "a", "", "AI agent to use: opencode, claude, cursor (default: opencode)")
	flags.StringVar(&cfg.Model, "model", "", "AI model to use (e.g., claude-sonnet-4, gpt-4)")
	flags.StringVar(&cfg.AgentMode, "agent-mode", "", "Agent mode/sub-agent to use (e.g., reviewer, planner)")
}

// RunLoop executes the main Ralph iteration loop.
func RunLoop(cfg *config.Config, promptText, promptName string, output io.Writer) error {
	completionSignal := "<promise>COMPLETE</promise>"
	writef := func(format string, args ...any) {
		_, _ = fmt.Fprintf(output, format, args...)
	}
	writeln := func(args ...any) {
		_, _ = fmt.Fprintln(output, args...)
	}

	// Replace placeholders in prompt
	promptText = strings.ReplaceAll(promptText, "<COMPLETION_SIGNAL>", completionSignal)

	// Get the configured agent
	agentInstance := agent.GetAgent(cfg.AgentName, cfg.Model, cfg.AgentMode)

	// Check if agent is available
	if !agentInstance.IsAvailable() {
		writef("Warning: %s agent not found in PATH, will continue anyway...\n", agentInstance.Name())
	}

	writef("Starting Ralph - Max iterations: %d\n", cfg.MaxIterations)
	writef("Using agent: %s\n", agentInstance.Name())

	for i := 1; i <= cfg.MaxIterations; i++ {
		writef("\n")
		writef("===============================================================\n")
		writef(" [%s] Iteration %d of %d (%s)\n", promptName, i, cfg.MaxIterations, time.Now().Format(time.RFC3339))
		writef("===============================================================\n")

		// Check if DEBUG mode (for testing)
		if os.Getenv("DEBUG") != "" {
			writeln(promptText)
			writef("\nAll planned tasks completed!\n")
			writef("Completed at iteration %d of %d\n", i, cfg.MaxIterations)

			return nil
		}

		// Execute the agent
		result, err := agentInstance.Execute(promptText, output)
		if err != nil {
			// Non-fatal error, continue to next iteration
			writef("Command execution warning: %v\n", err)
		}

		// Check for completion signal
		if hasCompletionSignal(result, completionSignal) {
			writef("\nAll planned tasks completed!\n")
			writef("Completed at iteration %d of %d\n", i, cfg.MaxIterations)

			return nil
		}

		writef("Iteration %d complete. Continuing...\n", i)
	}

	writef("\nReached max iterations (%d) without completing all planned tasks.\n", cfg.MaxIterations)

	return fmt.Errorf("max iterations reached")
}

func hasCompletionSignal(result, completionSignal string) bool {
	for line := range strings.SplitSeq(result, "\n") {
		if strings.TrimSpace(line) == completionSignal {
			return true
		}
	}

	return false
}
