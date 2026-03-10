package agent

import (
	"io"
)

// CursorAgent implements the Agent interface for the cursor CLI.
type CursorAgent struct {
	Model     string
	AgentMode string
}

// Execute runs cursor with the given prompt.
func (a *CursorAgent) Execute(prompt string, output io.Writer) (string, error) {
	// Cursor CLI uses: cursor [--model <model>] <prompt>
	args := []string{}
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}
	args = append(args, prompt)

	return executeAgentCommand("cursor", args, output, "cursor")
}

// Name returns the name of the agent.
func (a *CursorAgent) Name() string {
	return "cursor"
}

// IsAvailable checks if cursor is available in PATH.
func (a *CursorAgent) IsAvailable() bool {
	return isAgentAvailable("cursor")
}
