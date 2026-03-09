package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// InitSession represents one interactive run of ralph init.
type InitSession struct {
	OutputPath          string
	IsTTY               bool
	ExistingConfigFound bool
	Questions           []InitQuestion
	Answers             *InitAnswers
	Confirmed           bool
	Reader              *bufio.Reader
	Writer              io.Writer
}

// InitQuestion represents a single question in the interactive flow.
type InitQuestion struct {
	Key          string
	Prompt       string
	Type         string // "select", "input", "confirm"
	DefaultValue string
	Options      []string
	Required     bool
	Validator    func(string) error
}

// InitAnswers mirrors configuration fields for collection.
type InitAnswers struct {
	AgentName              string
	Model                  string
	AgentMode              string
	MaxIterations          int
	SpecsDir               string
	SpecsIndexFile         string
	ImplementationPlanName string
	PromptsDir             string
	NoLog                  bool
	LogFile                string
	LogTruncate            bool
}

// NewInitCommand creates the init command.
func NewInitCommand() *cobra.Command {
	var force bool
	var output string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Ralph configuration",
		Long:  `Interactive command to generate a ralph.toml configuration file.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			session := &InitSession{
				OutputPath: output,
				Answers:    &InitAnswers{},
				Reader:     bufio.NewReader(os.Stdin),
				Writer:     os.Stdout,
			}

			// TTY detection
			fileInfo, _ := os.Stdout.Stat()
			session.IsTTY = (fileInfo.Mode() & os.ModeCharDevice) != 0

			if !session.IsTTY {
				return fmt.Errorf("ralph init requires an interactive terminal")
			}

			// Default output path
			if session.OutputPath == "" {
				cwd, _ := os.Getwd()
				session.OutputPath = filepath.Join(cwd, "ralph.toml")
			}

			// Check for existing config
			if _, err := os.Stat(session.OutputPath); err == nil {
				session.ExistingConfigFound = true
				if !force {
					// Logic for overwrite confirmation would go here.
					// For now, we just proceed as the interactive flow will handle it,
					// or we can fail if we want to be strict before implementing questions.
					// Implementation of the full flow is in the next phase.
					return fmt.Errorf("configuration file already exists at %s (use --force to overwrite)", session.OutputPath)
				}
			}

			// Implementation of the full interactive flow is pending.
			_, _ = fmt.Fprintf(session.Writer, "Initializing Ralph configuration at %s\n", session.OutputPath)

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config without prompt")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Target file path (default: ./ralph.toml)")

	return cmd
}
